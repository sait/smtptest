package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

func main() {
	// --- Parámetros de línea de comando ---
	host := flag.String("host", "", "Servidor SMTP (ej: smtp.gmail.com)")
	port := flag.String("port", "587", "Puerto SMTP (ej: 25, 465, 587)")
	user := flag.String("user", "", "Usuario de autenticación SMTP")
	pswd := flag.String("pswd", "", "Contraseña SMTP")
	from := flag.String("from", "", "Correo remitente (From). Si se omite, se usa -user")
	to := flag.String("to", "", "Destinatario (ej: destino@ejemplo.com)")
	msgsubj := flag.String("msgsubj", "Prueba SMTP", "Asunto del correo")
	msgbody := flag.String("msgbody", "Este es un correo de prueba.", "Cuerpo del correo")
	insecure := flag.Bool("insecure", false, "Permitir autenticación sin TLS si el servidor no ofrece STARTTLS (credenciales en texto plano, no recomendado)")
	flag.Parse()

	// Logger con timestamp
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	logger.Println("=== Iniciando prueba de envío SMTP ===")

	// --- Validar parámetros obligatorios ---
	var faltantes []string
	if *host == "" {
		faltantes = append(faltantes, "-host")
	}
	if *user == "" {
		faltantes = append(faltantes, "-user")
	}
	if *pswd == "" {
		faltantes = append(faltantes, "-pswd")
	}
	if *to == "" {
		faltantes = append(faltantes, "-to")
	}
	if len(faltantes) > 0 {
		logger.Printf("[ERROR] Faltan parámetros obligatorios: %s", strings.Join(faltantes, ", "))
		fmt.Println("\nUso:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Si no se especifica -from, se usa -user como remitente
	if *from == "" {
		from = user
	}

	addr := net.JoinHostPort(*host, *port)
	logger.Printf("[INFO]  Host:        %s", *host)
	logger.Printf("[INFO]  Puerto:      %s", *port)
	logger.Printf("[INFO]  Dirección:   %s", addr)
	logger.Printf("[INFO]  Usuario:     %s", *user)
	logger.Printf("[INFO]  Remitente:   %s", *from)
	logger.Printf("[INFO]  Destinatario:%s", *to)
	logger.Printf("[INFO]  Asunto:      %s", *msgsubj)

	// --- Construir el mensaje RFC 2822 ---
	logger.Println("[PASO 1] Construyendo el mensaje de correo...")
	mensaje := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nDate: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		*from,
		*to,
		time.Now().Format(time.RFC1123Z),
		*msgsubj,
		*msgbody,
	)
	logger.Println("[PASO 1] Mensaje construido correctamente.")

	// --- Intentar conexión TCP ---
	logger.Printf("[PASO 2] Conectando a %s ...", addr)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		logger.Fatalf("[ERROR] No se pudo conectar al servidor: %v", err)
	}
	logger.Println("[PASO 2] Conexión TCP establecida.")
	conn.Close()

	// --- Seleccionar método de envío según puerto ---
	switch *port {
	case "465":
		enviarSSL(logger, *host, addr, *user, *pswd, *from, *to, mensaje)
	default:
		// Puerto 587 (STARTTLS) o 25 (sin TLS como fallback)
		enviarSTARTTLS(logger, *host, addr, *user, *pswd, *from, *to, mensaje, *insecure)
	}
}

// insecurePlainAuth es como smtp.PlainAuth pero sin la verificación de TLS,
// para servidores que aceptan AUTH PLAIN en texto plano (ej: Postal sin TLS).
type insecurePlainAuth struct {
	username, password string
}

func (a *insecurePlainAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	resp := []byte("\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *insecurePlainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, fmt.Errorf("respuesta inesperada del servidor durante AUTH PLAIN")
	}
	return nil, nil
}

// enviarSTARTTLS usa STARTTLS (puertos 587 / 25)
func enviarSTARTTLS(logger *log.Logger, host, addr, user, pswd, from, to, mensaje string, insecure bool) {
	logger.Println("[PASO 3] Iniciando cliente SMTP con STARTTLS...")

	c, err := smtp.Dial(addr)
	if err != nil {
		logger.Fatalf("[ERROR] smtp.Dial falló: %v", err)
	}
	defer c.Close()
	logger.Println("[PASO 3] Conexión SMTP establecida. Saludando al servidor...")

	if err = c.Hello("localhost"); err != nil {
		logger.Fatalf("[ERROR] EHLO/HELO falló: %v", err)
	}

	// Negociar STARTTLS si el servidor lo soporta
	tlsActivo, _ := c.Extension("STARTTLS")
	if tlsActivo {
		logger.Println("[PASO 4] Servidor soporta STARTTLS. Iniciando upgrade TLS...")
		tlsCfg := &tls.Config{ServerName: host}
		if err = c.StartTLS(tlsCfg); err != nil {
			logger.Fatalf("[ERROR] STARTTLS falló: %v", err)
		}
		logger.Println("[PASO 4] TLS activado correctamente.")
	} else if insecure {
		logger.Println("[PASO 4] AVISO: el servidor no ofrece STARTTLS. Continuando SIN CIFRAR (modo -insecure, las credenciales viajan en texto plano).")
	} else {
		logger.Fatalf("[ERROR] El servidor no ofrece STARTTLS. Use -insecure para autenticar sin cifrado (no recomendado) o habilite TLS en el servidor.")
	}

	// Autenticación
	logger.Printf("[PASO 5] Autenticando como %s ...", user)
	var auth smtp.Auth
	if tlsActivo {
		auth = smtp.PlainAuth("", user, pswd, host)
	} else {
		auth = &insecurePlainAuth{username: user, password: pswd}
	}
	if err = c.Auth(auth); err != nil {
		logger.Fatalf("[ERROR] Autenticación fallida: %v", err)
	}
	logger.Println("[PASO 5] Autenticación exitosa.")

	enviarMensaje(logger, c, from, to, []byte(mensaje))
}

// enviarSSL usa conexión directa TLS (puerto 465)
func enviarSSL(logger *log.Logger, host, addr, user, pswd, from, to, mensaje string) {
	logger.Println("[PASO 3] Iniciando conexión TLS directa (SSL, puerto 465)...")

	tlsCfg := &tls.Config{ServerName: host}
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		logger.Fatalf("[ERROR] tls.Dial falló: %v", err)
	}
	logger.Println("[PASO 3] Conexión TLS establecida.")

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		logger.Fatalf("[ERROR] No se pudo crear cliente SMTP sobre TLS: %v", err)
	}
	defer c.Close()

	logger.Printf("[PASO 4] Autenticando como %s ...", user)
	auth := smtp.PlainAuth("", user, pswd, host)
	if err = c.Auth(auth); err != nil {
		logger.Fatalf("[ERROR] Autenticación fallida: %v", err)
	}
	logger.Println("[PASO 4] Autenticación exitosa.")

	enviarMensaje(logger, c, from, to, []byte(mensaje))
}

// enviarMensaje realiza MAIL FROM / RCPT TO / DATA
func enviarMensaje(logger *log.Logger, c *smtp.Client, from, to string, msg []byte) {
	logger.Printf("[PASO 6] Estableciendo remitente: %s", from)
	if err := c.Mail(from); err != nil {
		logger.Fatalf("[ERROR] MAIL FROM falló: %v", err)
	}

	logger.Printf("[PASO 7] Estableciendo destinatario: %s", to)
	if err := c.Rcpt(to); err != nil {
		logger.Fatalf("[ERROR] RCPT TO falló: %v", err)
	}

	logger.Println("[PASO 8] Enviando cuerpo del mensaje (DATA)...")
	wc, err := c.Data()
	if err != nil {
		logger.Fatalf("[ERROR] Comando DATA falló: %v", err)
	}

	if _, err = wc.Write(msg); err != nil {
		logger.Fatalf("[ERROR] Escritura del mensaje falló: %v", err)
	}
	if err = wc.Close(); err != nil {
		logger.Fatalf("[ERROR] Cierre de DATA falló: %v", err)
	}

	logger.Println("[PASO 9] Enviando QUIT al servidor...")
	c.Quit()

	logger.Println("=== ✅ Correo enviado exitosamente ===")
}
