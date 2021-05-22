# Willkommen!

Das **Endpoint System (EPS)** bietet mehrere Server- und Client-Komponenten, die die Kommunikation im IRIS-Ökosystem verwalten und sichern. Insbesondere bietet das EPS zwei Kernkomponenten:

* Ein **Message-Broker / Mesh-Router-Dienst**, der Anfragen zwischen verschiedenen Akteuren im System weiterleitet und die gegenseitige Autorisierung und Authentifizierung sicherstellt.
* Ein **verteiltes Dienstverzeichnis**, das kryptografisch signierte Informationen über Akteure im System speichert und vom Message Broker für die Authentifizierung, die Dienstsuche und den Verbindungsaufbau verwendet wird.

Zusätzlich bietet es einen **TLS-Passthrough-Proxy-Dienst, der** eine direkte, Ende-zu-Ende-verschlüsselte Kommunikation zwischen Client-Endpunkten und Gesundheitsämtern ermöglicht.

## Erste Schritte

Für die ersten Schritte lesen Sie bitte die Anleitung " [Erste Schritte"]({{'getting-started'|href}}). Danach können Sie sich die [detaillierte EPS-Dokumentation]({{'eps.index'|href}}) ansehen. Wenn Sie den Proxy oder das Dienstverzeichnis ausführen möchten, können Sie sich die jeweilige [Proxy-Dokumentation]({{'proxy.index'|href}}) sowie die [Dienstverzeichnis-Dokumentation]({{'sd.index'|href}}) ansehen.

Wenn Sie auf ein Problem stoßen, [öffnen Sie](https://github.com/iris-connect/eps) bitte [ein Issue auf Github](https://github.com/iris-connect/eps), wo unsere Community Ihnen helfen kann.
