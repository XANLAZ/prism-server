# 🚀 OwpenGram Server

**A fully-featured (WIP) Telegram-compatible server, free for everyone to use.**

https://github.com/user-attachments/assets/3eee67ff-692f-455c-a1ca-444e25000865

---

## ✨ What is OwpenGram?

OwpenGram is an open-source Telegram-compatible server that will provide a complete MTProto implementation. Deploy your own messaging server and use it with comfort clients!

## 🎯 Current Features

- 💬 **Private Chats** - Full support for private messaging
- 👥 **Basic Groups** - Create and manage groups
- 📇 **Contacts** - Contact management and synchronization
- 🌐 **Web** - Web client support
- 📞 **1-to-1 Calls** - Voice and video calls

## 🚀 Quick Start

### Docker Installation

Get up and running in minutes with Docker! No manual setup required.

**1. Clone the repository**

```bash
git clone https://github.com/owpengram/owpengram-server.git
cd owpengram-server
```

**2. Start dependencies**

This starts MySQL, Redis, etcd, Kafka, and MinIO. Everything initializes automatically!

```bash
docker compose -f docker-compose-env.yaml up -d
```

**3. Start the application**

```bash
docker compose up -d
```

**Default verification code:** `12345` (change this for production!)

## 📱 Supported Clients

- 🤖 [Android Client](https://github.com/owpengram/owpengram-android-client)
- 💻 [Desktop Client](https://github.com/owpengram/owpengram-desktop-client)

## 💬 Community

- 📢 **Telegram Channel:** [@owpengram](https://t.me/owpengram)
- 💬 **Telegram Chat:** [Join the discussion](https://t.me/+sVB6Ymv70jEwNTAy)

For detailed documentation and advanced setup, check out the [original Teamgram Server repository](https://github.com/teamgram/teamgram-server).

## 📄 License

[Apache License 2.0](LICENSE)

---

## ⭐ Give us a Star!

If OwpenGram helps you, consider giving us a star on GitHub!
