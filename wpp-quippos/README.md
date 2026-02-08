# wpp-quippos

Servidor webhook para la API de WhatsApp de Meta.

## 📋 Descripción

Bot de WhatsApp con IA integrada (Google Gemini) que recibe mensajes y responde automáticamente.

**⚡ Para configurar rápido**: Lee [SETUP.md](./SETUP.md) - Guía de configuración mínima

## 🚀 Inicio Rápido

### 1. Configurar Variables de Entorno

Copia el archivo de ejemplo y configura tus valores:

```bash
copy .env.example .env
```

Edita `.env` y configura:
- `PORT`: Puerto donde correrá el servidor (default: 8080)
- `VERIFY_TOKEN`: Token secreto que usarás para verificar el webhook con Meta

### 2. Ejecutar el Servidor

```bash
go run main.go
```

El servidor iniciará en el puerto configurado (por defecto 8080).

### 3. Exponer con ngrok

En otra terminal, ejecuta ngrok para exponer tu servidor localmente:

```bash
ngrok http 8080
```

ngrok te dará una URL HTTPS pública (ejemplo: `https://abc123.ngrok.io`).

### 4. Configurar Webhook en Meta Developer Console

1. Ve a [Meta for Developers](https://developers.facebook.com/)
2. Selecciona tu aplicación de WhatsApp Business
3. Ve a **Configuración de Webhooks** (WhatsApp > Configuration > Webhooks)
4. Haz clic en **Editar** o **Configurar webhooks**
5. Ingresa:
   - **URL de devolución de llamada**: `https://TU-URL-DE-NGROK.ngrok.io/webhook`
   - **Token de verificación**: El mismo valor que configuraste en `VERIFY_TOKEN`
6. Haz clic en **Verificar y guardar**
7. Suscríbete a los eventos que desees recibir (messages, message_status, etc.)

## 📡 Endpoints

### `GET /webhook`
Endpoint de verificación del webhook (requerido por Meta).

**Parámetros:**
- `hub.mode`: debe ser "subscribe"
- `hub.verify_token`: tu token de verificación
- `hub.challenge`: valor que debe devolverse para confirmar

**Respuesta exitosa:** Devuelve el `hub.challenge`

### `POST /webhook`
Endpoint que recibe los eventos de WhatsApp.

**Comportamiento actual:**
- Imprime el JSON completo en consola con formato legible
- Extrae y muestra información relevante (tipo de evento, número de mensajes, etc.)
- Devuelve 200 OK para confirmar recepción

## 🔍 Ejemplo de Output

Cuando recibes un mensaje, verás algo como:

```
===================================================
📨 POST /webhook - Evento recibido [2026-02-06 19:58:00]
===================================================
{
  "object": "whatsapp_business_account",
  "entry": [
    {
      "id": "123456789",
      "changes": [
        {
          "value": {
            "messaging_product": "whatsapp",
            "metadata": { ... },
            "messages": [
              {
                "from": "5491112345678",
                "id": "wamid.xxx",
                "timestamp": "1234567890",
                "text": {
                  "body": "Hola!"
                },
                "type": "text"
              }
            ]
          },
          "field": "messages"
        }
      ]
    }
  ]
}
===================================================
📌 Tipo de objeto: whatsapp_business_account
📌 Número de entradas: 1
💬 Se recibieron 1 mensaje(s)
---------------------------------------------------
```

## 🛠️ Compilación

Para compilar un ejecutable:

```bash
go build -o wpp-quippos.exe
```

Luego ejecuta:

```bash
./wpp-quippos.exe
```

## 📝 Estructura del Proyecto

```
wpp-quippos/
├── main.go           # Servidor HTTP con lógica de webhook
├── go.mod            # Definición de módulo Go
├── .env.example      # Ejemplo de variables de entorno
├── .env              # Variables de entorno (no incluido en git)
└── README.md         # Este archivo
```

## 🔐 Seguridad

- **NUNCA** subas tu archivo `.env` a git
- Cambia el `VERIFY_TOKEN` a un valor seguro y aleatorio
- En producción, considera agregar autenticación adicional
- Valida siempre las firmas de las solicitudes de Meta

## 🚧 Próximos Pasos

Este servidor está configurado para desarrollo inicial. Para producción, considera:

- [ ] Validar firma de las solicitudes (X-Hub-Signature-256)
- [ ] Almacenar mensajes en base de datos
- [ ] Implementar lógica de respuesta automática
- [ ] Agregar manejo robusto de errores
- [ ] Implementar rate limiting
- [ ] Configurar logging estructurado

## 📚 Recursos

- [Documentación oficial de WhatsApp Business API](https://developers.facebook.com/docs/whatsapp/)
- [Webhooks de WhatsApp](https://developers.facebook.com/docs/whatsapp/webhooks/)
- [ngrok Documentation](https://ngrok.com/docs)
