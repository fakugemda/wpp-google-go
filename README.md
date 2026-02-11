<div align="center">

# 📧✨ MailBot AI

### *Asistente Inteligente Multi-Platforma para Gestión de Emails*

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2-00ADD8?style=for-the-badge)](https://gofiber.io/)
[![Gemini](https://img.shields.io/badge/Gemini-AI-4285F4?style=for-the-badge&logo=google)](https://ai.google.dev/)

**Gestiona tus correos electrónicos desde WhatsApp y Discord usando Inteligencia Artificial**

---

</div>

## 🎯 ¿Qué es MailBot AI?

MailBot AI es un asistente inteligente que te permite **redactar y enviar emails** directamente desde **WhatsApp** o **Discord** usando **Google Gemini AI**. El bot entiende tus instrucciones en lenguaje natural, genera borradores profesionales y los envía a través de tu cuenta de Gmail.

## ✨ Funcionalidades

### 🤖 Procesamiento Inteligente

- **IA con Gemini**: Procesa mensajes en lenguaje natural y genera respuestas contextuales
- **Reconocimiento de Intenciones**: Detecta automáticamente si quieres enviar un email, chatear o confirmar acciones
- **Gestión de Contactos**: Resuelve nombres a emails automáticamente desde tu agenda

### 📱 Multi-Platforma

- **WhatsApp**: Recibe y responde mensajes vía Meta Cloud API
- **Discord**: Soporte completo para mensajes directos (DMs)
- **Webhook**: Endpoints REST para integraciones externas

### 📧 Gestión de Emails

- **Creación de Borradores**: Genera emails profesionales con asunto y contenido
- **Envío Directo**: Confirma y envía correos con un simple comando
- **Validación Automática**: Verifica formatos de email y detecta contenido HTML
- **Adjuntos de Imágenes**: Envía imágenes recibidas por WhatsApp como adjuntos en los emails
- **Soporte UTF-8 Completo**: Asuntos y contenido con tildes, eñes y emojis se muestran correctamente

## 🏗️ Arquitectura

```
wpp-google-go/
├── 📁 auth/          # Autenticación OAuth2 con Google
├── 📁 config/        # Configuración y carga de contactos
├── 📁 constants/     # Constantes del proyecto
├── 📁 controller/    # Controladores HTTP (WhatsApp)
├── 📁 models/        # Estructuras de datos y templates
├── 📁 routes/        # Definición de rutas API
├── 📁 service/       # Lógica de negocio
│   ├── ai.go         # Integración con Gemini AI
│   ├── discord.go    # Servicio de Discord
│   ├── gmail.go      # Gestión de Gmail (con soporte UTF-8 y adjuntos)
│   ├── meta.go       # API de WhatsApp
│   └── meta_media.go # Descarga de imágenes desde Meta/WhatsApp
└── 📁 utils/         # Utilidades genéricas (Singleton)
```

## 🚀 Inicio Rápido

### Prerrequisitos

- Go 1.25+
- Cuenta de Google con Gmail habilitado
- Token de Discord Bot (opcional)
- Credenciales de Meta WhatsApp API (opcional)
- API Key de Google Gemini

### Configuración

1. **Clona el repositorio**

```bash
git clone <repo-url>
cd wpp-google-go
```

2. **Instala dependencias**

```bash
go mod download
```

3. **Configura variables de entorno**

```bash
# Gmail
export GOOGLE_CREDENTIALS="..." # JSON de credenciales OAuth2
export GOOGLE_TOKEN="..."       # Token OAuth2 (opcional)

# Gemini AI
export GEMINI_API_KEY="..."

# WhatsApp (opcional)
export META_TOKEN="..."
export META_PHONE_ID="..."
export META_VERIFY_TOKEN="..."

# Discord (opcional)
export DISCORD_TOKEN="..."

# Contactos
export CONTACTS_JSON='{"Nombre": "email@example.com"}' # JSON de contactos
```

4. **Ejecuta la aplicación**

```bash
go run main.go
```

## 📖 Uso

### Desde WhatsApp

#### Envío de Emails con Texto

1. Envía un mensaje al bot: *"Redacta un email a Juan sobre la reunión de mañana"*
2. El bot genera un borrador y te lo muestra
3. Responde **"sí"** o **"envíalo"** para confirmar el envío

#### Envío de Emails con Imágenes 📷

1. **Opción A - Imagen con Caption:**

   - Envía una imagen a WhatsApp con un texto (caption): *"Mándasela a mamá y dile que ya terminé"*
   - El bot descarga la imagen, procesa el texto y genera el borrador con la imagen adjunta
   - Confirma con **"sí"** para enviar
2. **Opción B - Imagen sin Caption:**

   - Envía solo la imagen (sin texto)
   - El bot te pregunta qué hacer con ella
   - Responde con instrucciones: *"Mándasela a mi jefe"*
   - El bot genera el borrador con la imagen adjunta
   - Confirma con **"sí"** para enviar

### Desde Discord

1. Abre un DM con el bot
2. Escribe tu solicitud en lenguaje natural
3. El bot procesa y responde automáticamente

### Comandos Especiales

- `cancelar` - Cancela operaciones pendientes y limpia el caché de imágenes
- `confirmar` / `sí` - Confirma el envío de un borrador

### Ejemplos de Uso

#### Ejemplo 1: Email Simple

```
Tú: "Redacta un email a juan@example.com sobre la reunión de mañana"
Bot: *Borrador IA Creado*
     *Para:* juan@example.com
     *Asunto:* Reunión de mañana
     [Contenido del email...]
     _¿Lo envío? (Responde Sí)_
Tú: "sí"
Bot: 🚀 Correo enviado exitosamente!
```

#### Ejemplo 2: Email con Imagen y Caption

```
Tú: [Envías una foto con caption: "Mándasela a mamá y dile que ya terminé"]
Bot: *Borrador IA Creado*
     *Para:* mamá@example.com
     *Asunto:* Actualización
     [Contenido del email...]
     📎 [Imagen adjunta]
     _¿Lo envío? (Responde Sí)_
Tú: "sí"
Bot: 🚀 Correo enviado exitosamente!
```

#### Ejemplo 3: Imagen sin Caption

```
Tú: [Envías solo una foto]
Bot: 📷 Foto recibida. ¿Qué quieres que haga con ella? (Ej: 'Mándasela a mamá')
Tú: "Mándasela a mi jefe con el asunto 'Proyecto terminado'"
Bot: *Borrador IA Creado*
     *Para:* jefe@example.com
     *Asunto:* Proyecto terminado
     [Contenido del email...]
     📎 [Imagen adjunta]
     _¿Lo envío? (Responde Sí)_
```

## 🔧 Tecnologías

- **Go 1.25** - Lenguaje principal
- **Fiber v2** - Framework web
- **Google Gemini AI** - Procesamiento de lenguaje natural
- **Gmail API** - Gestión de correos
- **DiscordGo** - Integración con Discord
- **Meta Cloud API** - Integración con WhatsApp

## 📝 Estructura de Datos

### AIResponse

```go
type AIResponse struct {
    Type    string // "CHAT" | "EMAIL_DRAFT" | "CONFIRM_SEND"
    Content string // Contenido del mensaje o email
    To      string // Email destinatario
    Subject string // Asunto del email
}
```

## 🎨 Patrones Implementados

- **Singleton**: Utils, Repository pattern
- **Dependency Injection**: Controllers y Services
- **Separation of Concerns**: Capas bien definidas (Controller → Service → Repository)
- **Caché con TTL**: Sistema de caché temporal para imágenes con limpieza automática (30 min TTL)
- **Thread-Safe**: Protección concurrente con mutex para operaciones sobre el caché

## 🔐 Características Técnicas

### Gestión de Imágenes

- **Descarga Automática**: Las imágenes se descargan automáticamente al recibirlas
- **Caché Temporal**: Las imágenes se almacenan en memoria con timestamp
- **Limpieza Automática**: Sistema de limpieza que elimina imágenes expiradas cada 2 minutos
- **TTL Configurable**: Las imágenes expiran después de 30 minutos sin uso
- **Detección de Tipo**: Identificación automática del tipo MIME y extensión de archivo

### Codificación UTF-8

- **Asuntos Blindados**: Codificación RFC 2047 para asuntos con caracteres especiales
- **Cuerpo Base64**: Codificación Base64 del cuerpo cuando hay adjuntos
- **Soporte Completo**: Tildes, eñes, emojis y caracteres especiales funcionan correctamente

## 📄 Licencia

Este proyecto es privado y confidencial.

---

<div align="center">

**Desarrollado con ❤️ usando Go y Gemini AI**

[Reportar Bug](https://github.com/fakugemda/wpp-google-go/issues) · [Solicitar Feature](https://github.com/fakugemda/wpp-google-go/issues)

</div>
