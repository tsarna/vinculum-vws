# Vinculum WebSockets

"The [vinculum is the] processing device at the core of every Borg vessel.
It interconnects the minds of all the drones."
   -- Seven of Nine (In Voyager episode "Infinite Regress")

VWS (Vinculum WebSockets) is a simple, WebSocket based publish/subscribe protocol on top of WebSockets, with MQTT-esque fucntionality. This project, an add on to vinculum-bus, lets you expose an EventBus to clients over websockets using the server component, or to access such a server using the client component.

## **WebSocket Server**

Expose your EventBus over WebSockets for real-time web applications:
- **Real-time event streaming** to web clients
- **Bidirectional communication** (subscribe + publish)
- **Flexible authorization policies**
- **Built-in metrics** and connection management
- **Message transformations** and filtering

📖 **[WebSocket Server Documentation](server/README.md)**

## 🔌 **WebSocket Client**

Connect to Vinculum WebSocket servers from Go applications:
- **Auto-reconnection** with exponential backoff
- **Subscription management** and persistence
- **Thread-safe** operations
- **Builder pattern** for easy configuration

📖 **[WebSocket Client Documentation](client/README.md)**

## 📋 **Protocol**

Both components implement the Vinculum WebSocket Protocol:
- **JSON-based** with compact message format
- **MQTT-style** topic patterns
- **Request/response** correlation
- **Error handling** and acknowledgments

The protocol should be easy to implement in other languages, such as JavaScript.

📖 **[Protocol Specification](PROTOCOL.md)**

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Vinculum** (Latin: "bond" or "link") - connecting your application components with reliable, observable messaging.
