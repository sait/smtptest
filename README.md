# smtptest
smtptest Test send message using smtp


# To test
- download smtptest.zip
- unzip it
- in cmd run:

```
smtptest.exe -host smtp.gmail.com -port 465 -user tucorreo@gmail.com -pswd "tu_app_password" -to destino@ejemplo.com -msgsubj "Prueba de conexión"   -msgbody "Hola, esto es una prueba SMTP."
```

# Flags

| Flag | Obligatorio | Descripción |
|---|---|---|
| `-host` | Sí | Servidor SMTP (ej: smtp.gmail.com) |
| `-port` | No (default `587`) | Puerto SMTP (25, 465, 587, etc.) |
| `-user` | Sí | Usuario de autenticación SMTP |
| `-pswd` | Sí | Contraseña SMTP |
| `-from` | No (default: valor de `-user`) | Correo remitente (From). Útil cuando `-user` es una credencial que no es un correo (ej. Postal) |
| `-to` | Sí | Destinatario |
| `-msgsubj` | No | Asunto del correo |
| `-msgbody` | No | Cuerpo del correo |
| `-insecure` | No | Permite autenticarse sin TLS si el servidor no ofrece STARTTLS (credenciales en texto plano). **No recomendado**, solo para pruebas puntuales contra servidores sin TLS configurado |

# Ejemplo con -insecure

Si el servidor no ofrece STARTTLS (por ejemplo, un Postal sin TLS habilitado), el programa falla por defecto para evitar enviar la contraseña en texto plano. Para forzar la prueba de todas formas:

```
smtptest -host postal.ejemplo.com -port 2525 -user "org/mailserver" -pswd "xxxx" -from noreply@ejemplo.com -to destino@ejemplo.com -insecure
```
