# IRIS Endpunkt-Server - Design

Der IRIS Endpunkt-Server (Arbeitsname) handhabt die Kommunikation zwischen Drittanbietern und Gesundheitsämtern einerseits sowie dem zentralen IRIS-Gateway andererseits (im dezentralen Ansatz handhabt er direkt die Kommunikation zwischen Gesundheitsämtern und Drittanbietern). Ziel des Servers ist es, zwischen den Kommunikationspartnern eine sichere Kommunikation zu ermöglichen und für diesed die technische Ausgestaltung der Kommunikation weitgehend zu abstrahieren. Hierdurch soll auch das Potential für endpunktseitige Fehler bei der Anbindung an das IRIS-System minimiert werden.

## Design-Ziele

* Einfaches Deployment des Servers mit möglichst wenig Abhängigkeiten
* Automatisierung aller Sicherheitsaspekte der Kommunikation (Verschlüsselung, Signierung, Schlüsselmanagement)
* Standardkonforme und einfache lokale Anbindung (z.B. via REST, JSON-RPC etc.)
* Hohes Verarbeitungspotential ohne notwendige horizontale Skalierung

## Umsetzung

Die Umsetzung soll in `Golang` erfolgen. `Golang` ist eine moderne, stark typisierte Sprache die insbesondere für die Implementierung von Internet-Diensten sehr gut geeignet ist und zudem hervorragende Unterstützung für moderne kryptographische Standards bietet. `Golang`-Programme werden zu statischen Binärdateien kompiliert, die praktisch ohne externe Abhängigkeiten ausgeführt werden können. In die Sprache integrierte Methoden zur Parallelisierung machen den Einsatz für die Server-Programmierung zudem sehr attraktiv. Durch automatisiertes Speicher-Management lassen sich zudem viele Fehlerquellen und damit Sicherheitsrisiken die sich typischerweise in Systemsprachen (C, C++) ergeben eliminieren.

## Funktionalität

Der Server soll sowohl eine interne wie auch externe API bieten. Die interne API ist zur Nutzung durch Gesundheitsämter oder Drittanbieter innerhalb ihrer eigenen Infrastruktur vorgesehen. Die externe API hingegen ist zur Nutzung durch andere Akteure im System vorgesehen (beispielsweise den zentralen IRIS-Gateway), und ermöglicht eine bidirektionale Kommunikation mit diesem. Beide APIs können über verschiedene Schnittstellen realisiert werden, beispielsweise JSON-RPC oder REST. Gegebenenfalls kann auch direkt eine auf "Message Passing" angepasste Technologie genutzt werden, z.B. bidirektionales gRPC über TLS.

### Externe API (Anbieter)

Die externe API bietet folgende Funktionalität:

* **ProvidesService(service Service)** → APIResponse(status <OK, ERR>, {OK: Service, ERR: Error})
* **Services()** → APIResponse(status <OK, ERR>, {OK: list<Service>, ERR: Error})
* **SubmitRequest(request MessageRequest)** → APIResponse(status <OK, ERR>)
* **GetResponse(requestID uuid)** → APIResponse(status <OK, ERR>, {OK: MessageResponse, ERR: Error}>

Im Falle der Realisierung über gRPC oder JSON-RPC stellt jede Funktion einen `Message`-Typ dar, im Falle von REST jeweils einen einzelnen Endpunkt. gRPC kann zusätzlich `MessageResponse` Objekte aktiv an die Gegenstelle senden, im REST/JSON-RPC Ansatz ist hierzu ein Polling der Gegenstelle oder ein anderer HTTPs-kompatibler Mechanismus (z.B. Web-Sockets, HTTP/3.0 Server Push) notwendig.

### Interne API (Anbieter)

* **GetRequests(map<string, any> filters, int limit, int offset)** → *APIResponse(status <OK, ERR>, data {OK: list<MessageRequest>, ERR: Error})*
* **GetRequest(id uuid)** → *APIResponse(<OK, ERR>, {OK: Request, ERR: Error})*
* **SetResponse(id uuid, response MessageResponse)** → *Response(<OK, ERR>)*

Zusätzlich kann auch ein `Hook`-Mechanismus implementiert werden, der einkommende `MessageRequest` Objekte automatisch an einen interen API-Endpunkt weiterleitet. 

### Interne API (GA)

* **SubmitMessage(message Message)** → *APIResponse(status <OK, ERR>)*
* **GetMessages(map<string, any> filters, int limit, int offset)** → *APIResponse(status <OK, ERR>, data {OK: list<Message>, ERR: Error})*
* **GetMessage(id uuid)** → *APIResponse(<OK, ERR>, {OK: Message, ERR: Error})*
* **UpdateMessage(id uuid, data map<string, any>)** → *Response(<OK, ERR>, {OK: Message, ERR: Error})*
* **DeleteMessage(id uuid, data map<string, any>)** → *Response(<OK, ERR>, {OK: nil, ERR: Error})*

Wie oben können eingehende Nachrichten über einen `Hook`-Mechanismus implementiert werden, der einkommende `MessageRequest` Objekte automatisch an einen interen API-Endpunkt weiterleitet. 

### Persistenz von Nachrichten

Unter der Annahme, dass `MessageRequest` und `MessageResponse` Objekte nur kurzzeitig im System vorgehalten werden, kann auf eine Persistierung gänzlich verzichtet werden. Wenn eine höhere Ausfallsicherheit sowie Robustheit gegen Abstürze gewünscht ist, kann diese z.B. über verschlüsseltes Write-Ahead-Logging (WAL) realisiert werden. Hierbei werden eingehende Nachrichten in einer Logdatei verschlüsselt abgelegt (nur wenn sie nicht direkt zustellbar sind) und können im Falle eines unerwarteten Neustarts des System von dort wieder eingelesen werden. WAL-Logs werden hierbei im "append only" Modus atomar mit Datenpaketen beschrieben und regelmäßig gelöscht. Generell kann die Zustellung bei entsprechend vorhandener Anbindung der Endsystem jedoch überwiegend synchron erfolgen, d.h. nicht zugestellbare Nachrichten werden abgelehnt und Fehler direkt an den entsprechenden Endpunkt zurückgegeben, die Fehlerbehandlung kann dann dort erfolgen. Generell sollte der Zeitraum der Zwischenspeicherung möglichst gering sein um das Risiko des Datenverlustes zu minimieren, eine Datenhaltung im Speicher ist daher gegenüber einer Persistierung z.B. in einer Datenbank zu bevorzugen.

In jedem Fall muss die Lebenszeit von Nachrichten im System beschränkt werden. Nach einer vordefinierten Zeit nicht abgeholte oder zustellbare Nachrichten sollen gelöscht werden, die Gegenseite soll hierüber ggf. eine Fehlerbenachrichtung erhalten.

## Umsetzung

Generell soll ein Server bereitgestellt werden, der möglichst einfach deployed werden kann. Hierzu soll ein einzelnes Binärprogramm ausgeliefert werden, welches über typisierte Einstellungen konfiguriert werden kann. Er soll direkt ausführbar sein und keine externen Abhängigkeiten (Datenbanken etc.) haben:

```bash
# Start des Endpunkt-Servers
iris-api
```

Der Server soll hierbei zwei TCP-Ports für die Anbindung der internen sowie externen Schnittstelle öffnen. Die externe Schnittstelle kann optional hinter einem TCP-LoadBalancer / Proxy (z.b. `haproxy`) betrieben werden. Die externe Server-Schnitstelle integriert Rate-Limiting und IP-Whitelisting auf TCP-Ebene und ist hierdurch im Idealfall nur durch vertrauenswürdige Gegenseiten erreichbar. Gegebenfalls sollte dies über zusätzliche technische Maßnahmen unabhängig abgesichert werden (z.B. Kernel-Firewalling auf IP-Ebene).

Es soll weiterhin über die Gegenstelle des Servers ein automatisierter Test zur Erreichbarkeit der internen Schnittstelle erfolgen. Ist diese von der öffentlichen IP-Adresse der Gegenstelle zu erreichen soll der Server automatisch terminiert werden. Die interne Schnittstelle soll zusätzlich über einen Token-Authentifizierungsmechanismus abgesichert.

### Benötigte Konfigurationsdaten

* Lokales TLS-Zertifikat
* CA-Zertifikat / Gegenseitiges TLS-Zertifikat
* Server-Name (nur für HTTPs-basierten Ansatz; kann ggf. aus lokalem Zertifikat abgeleitet werden)

Zusätzlich können weitere Konfigurationseinstellungen verfügbar gemacht werden, z.B. für IP-Whitelisting, Daten-Persistierung, lokale Weiterleitung von Daten (z.B. über API Hooks). Nicht sensible Konfigurationsdaten können aus Dateien geladen werden, sensible Daten sollten ggf. über einen externen Secrets-Management Mechanismus bereitgestellt oder manuell über die Kommandozeile injiziert werden können.

### Lokale Anbindung des Systems

Um Daten in das System zu senden können API-Endpunkte (REST oder JSON-RPC) bereitgestellt werden. Diese können zusätzlich über einen weiteren Authentifizierungs-Mechanismus gesichert werden (z.B. tokenbasierte Authentifizierung) und ebenfalls über (einfaches) TLS mit selbstsignierten Zertifikaten verschlüsselt bereitgestellt werden.

Zusätzlich können `Hooks` definiert werden, über die eingehende Nachrichten synchron über eine HTTPs `POST` Anfrage an das interne System weitergeleitet werden. Die Implementierung und Absicherung dieser Endpunkte obliegt dem Systembetreiber (die Angabe eines statischen Bearer-Tokens soll möglich sein). Beispiel:

```
POST /hook HTTP/1.1
Authentication: bearer a5ca5cd....
Content-Type: application/json; gzip
...

{
	"type": "cwa_contact_diary",
	"id": "a5fca...",
	"data": {
		//...
	}
}

```

## Kryptographisches Konzept

Der API-Server verschlüsselt Daten bei der Übertragung mittels Transportverschlüsselung. Hierbei wird gegenseitige TLS-Authentifizierung genutzt um sowohl Absender als auch Empfänger zu authentifizieren. Hierdurch lässt sich über standardisierte, zertifikatsbasierte Verschlüsselung sicherstellen, das auf dem Transportweg keine Daten durch einen Dritten manipuliert oder entschlüsselt werden können.

### Ende-zu-Ende Verschlüsselung

Immer wieder kommt bei der Gestaltung von Kommunikationssystemen auch das Konzept der "Ende-zu-Ende Verschlüsselung" auf. Besondere Relevanz hat es in Systemen, bei denen die Kommunikation zwischen zwei Endpunkten - z.B. einem Gesundheitsamt und einem Drittanbieter - durch einen Dritten - z.B. den IRIS-Gateway - vermittelt wird. Da keine direkte, synchrone Kommunikation zwischen den Endenpunkte erfolgt muss der vermittelnde Dritte Daten der Beteiligten zwischenspeichern oder zumindest temporär lokal zur Weiterleitung entschlüsseln. Um diese Daten bei der Zwischenspeicherung vor unberechtigtem Zugriff zu schützen, nutzen die Kommunikationspartner auf ihren Endpunkte neben der Transportverschlüsselung - die nur die Kommunikation zur vermittelnden Stelle absichert - eine weitere Verschlüsselung auf Applikationsebene. Für diese werden initiale kryptographische Schlüssel (im Idealfall) über einen separaten Vertrauensmechanismus ausgetauscht und durch verschiedene Techniken (kryptographische Ratschen, Diffie-Hellman Schlüsselaustausch) im Rahmen der Kommunikation häufig geändert. Hierüber soll sichergestellt werden, dass ein Angreifer der Zugang zu den Daten des vermittelnden Dritten erhält (sowie ggf. der vermittelnde Dritte selbst) nicht in der Lage ist, die dort zwischengespeicherten Daten zu entschlüsseln (sowie einiger weiterer Garantien, siehe unten).

Eine solche zusätzliche Verschlüsselung auf Applikationsebene ist aus Sicht des Autors jedoch nur sinnvoll wenn Daten über mehrere, nicht vertrauenswürdige Zwischenpunkte ausgetauscht werden müssen. Für Daten, die über eine transportverschlüsselte Verbindung direkt zwischen zwei Endpunkten ausgetauscht werden, wobei die Endpunkte selbst die Entschlüsselung vornehmen, ist eine zusätzliche Verschlüsselung auf Applikationsebene nicht sinnvoll oder birgt zumindest aus Sicherheitssicht keinen erheblichen zusätzlichen Nutzen (da Transport- und Applikationsverschlüsselung von den gleichen Systemen durchgeführt werden und dementsprechend alle hierzu notwendigen Schlüssel dort vorliegen müssen).

### Mögliches Ende-zu-Ende Konzept (optional und nur für Bedarfsfalll)

Generell verfügt jeder Kommunikationspartner im System über asymmetrische Schlüsselpaare zum Signieren und Verschlüsseln von Daten. Die Authentizität dieser Schlüssel wird durch ein oder mehrere Root-Zertifikate sichergestellt, mit denen direkt oder mittelbar Schlüssel signiert werden.

Die statische Verschlüsselung von Daten mit diesen Schlüseln (im Folgenden als statische Schlüssel bezeichnet) birgt jedoch das Risiko, dass ein Angreifer der Zugang zu einem privaten statischen Schlüssel erlangt und die Kommunikation im System überwacht in der Lage ist, auch rückwirkend Nachrichten zu entschlüsseln.

In der Praxis wurden verschiedene Methoden entwickelt um diese rückwirkende Entschlüsselung von Nachrichten (sowie die Entschlüsselung zukünftiger Nachrichten) zu erschweren oder gänzlich zu unterbinden. Hierbei wird der statische Schlüssel lediglich einmalig zur Etablierung eines vertrauenswürdigen Grundschlüssels genutzt, anschließend werden mithilfe eines Schlüsselaustauschproktolls (key agreement protocol) temporäre Schlüssel generiert, mit denen die tatsächliche Verschlüsselung erfolgt. Diese temporären Schlüssel werden so oft wie möglich gewechselt (im Idealfall für jede ausgetauschte Nachricht) und nach Verwendung zerstört. Dies macht es einem Angreifer schwerer oder unmöglich, durch Erlangung einzelner Schlüssel mehrere Nachrichten entschlüsseln zu können. Insbesondere im Messaging-Bereich haben sich diese Protokolle durchgesetzt, der aktuelle "Standard" ist das "Double Ratchet Protocol", welches Bestandteil des Nachrichtenprotokolls des "Signal" Messengers ist.

Für den Datenaustausch über ein zentralisiertes System sowie die Zwischenspeicherung von Daten in einem nicht voll vertrauenswürdigen System bieten diese Protokolle verschiedene Vorteile:

* Da statische Schlüssel nur zu Beginn der Kommunikation zwischen zwei Teilnehmern im System genutzt werden, führt die Erlangung dieser Schlüssel durch einen Angreifer nur sehr eingeschränkt oder gar nicht zu der Fähigkeit, abgefangene Nachrichten zu entschlüsseln.
* Da für jede Nachricht im System ein neuer temporärer Schlüssel verwendet wird und vergangene sowie zukünftige temporäre Schlüssel sich nicht (oder nur sehr beschränkt) von erlangten aktuellen temporären Schlüsseln ableiten lassen kann ein Angreifer erlangte temporäre Schlüssel nicht dazu einsetzen, vergangene oder zukünftige Nachrichten zu entschlüsseln.

Zur Umsetzung können bestehende Implementierungen des "Double Ratchet" Protokolls genutzt werden. Alternativ kann zunächst eine einfache ECDH-basierte Schlüsselgenerierung verwendet werden. Da im Normalfall die Kommunikation in Form von `Request`-`Response` Zyklen erfolgt ist der Sicherheitsverlust hierbei minimal, da ECDH-Schlüssel in jedem Zyklus ausgetauscht werden können. (ein "Double Ratchet" Ansatz ist lediglich vorteilhaft, wenn ein Kommunikationspartner einseitig eine große Anzahl sequenzieller Nachrichten schickt. In diesem Fall würde bei einer reinen ECDH-basierten Schlüsselgenerierung jeweils der gleiche Schlüssel für alle ausgehenden Nachrichten verwendet).

Die Verwaltung der jeweiligen aktuellen Schlüsselpaare muss lokal durch die Endpunkt-Server erfolgen.

### Signierung von Daten

Gegebenenfalls kann eine Signierung von `MessageRequest` und `MessageResponse` Objekten einseitig oder für beide Kommunikationspartner sinnvoll sein. Dies ist insbesondere der Fall, wenn Signaturschlüssel nicht vom Endpunkt-Server verwaltet werden sondern Anfragen von einem vertrauenswürdigeren, internen System generiert werden sollen und interne Systeme des Kommunikationspartners die Signaturen ebenfalls unabhängig vom Gateway-Server überprüfen sollen. Das hier vorgestellte System bietet die Möglichkeit, Anfragen und Antworten optional zu signieren. Eine Generierung und Prüfung von Signaturen kann im Gateway-Server erfolgen, ist dort jedoch nicht unbedingt sinnvoll, da zwischen den Endpunkten bereits eine Authentifizierung erfolgt. Jedoch können Signaturen zur gegenseitigen internen Prüfung bei den Kommunikationspartnern als zusätzliche Sicherheitsmaßnahme durchaus sinvoll sein. Hierüber kann beispielsweise vermieden werden, dass ein Angreifer dem es gelingt einen Gateway-Server zu ersetzen oder Daten über den `Hooks` Mechanismus an interne Systeme zu senden nicht in der Lage ist, valide Anfragen zu generieren.

Um einen Signaturmechanismus nutzen zu können ist wiederum ein Vertrauensmechanismus notwendig, der die Vertrauenswürdigkeit und Validität von Signaturschlüsseln überprüfen kann.

### Datentypen

Die folgenden Abschnitte beschreiben die vom System verarbeiteten Datentypen.

#### Terminologie

Typen werden nach Namen angegeben (Golang-Konvention), z.B. “limit int”. Geschweifte Klammern `{}` zeigen konditionale Rückgabewerte in Bezug auf Werte vorheriger ENUM-Parameter. Größer/Kleiner-Klammern `<>` bezeichnen mögliche ENUM-Werte sowie Parameter komplexer Typen wie z.B. Mappings.

**Message**: (soll in Datenbank persistiert werden)

* **id uuid**: Eindeutiger Identifer (auch zur Übermittlung an Anbieter-API)
* **created_at timestamp**: Erzeugungsdatum
* **status <INITIALIZED, SUBMITTED, REJECTED, ANSWERED, SUCCEEDED, FAILED, LOST>**: Aktueller Status des Objekts 
* **synchronous bool**: Falls gesetzt wird die Nachricht synchron übermittelt.
* **request MessageRequest**: Anfrage mit Methode, Parameter und Signatur
* **timeout \*bool**: Timeout für die Bearbeitung der Nachricht (danach wird sie als verloren gemeldet und gelöscht).
* **original_id \*uuid**: Identifer der Hauptnachricht aus welcher diese Nachricht resultierte (falls mehrere Nachrichten generiert werden)
* **updated_at \*timestamp**: Aktualisierungsdatum (falls zutreffend)
* **error \*Error**: Fehlerinformation (falls zutreffend)
* **data \*map<string, any>**: Beschreibende Daten zu dem Message-Objekt, z.B. Anmerkungen von einem GA-Mitarbeiter, für die interne Verwendung (falls zutreffend)
* **processing_result \*map<string, any>**: Beschreibende Daten zu dem Ergebnis der Verarbeitung im System, z.B. Anzahl integrierter Datensätze (falls zutreffend)
* **deleted_at \*timestamp**: Löschdatum (falls zutreffend)
* **response \*MessageResponse**: Antwort zur Anfrage (falls zutreffend)

**MessageResponse**:

* **status <OK, ERR, PENDING>**: Status der Antwort
* **data \*any**: Übermittelte Daten (falls zutreffend)
* **error \*Error**: Fehlerinformation (falls zutreffend)
* **signature \*Signature**: Signatur mit Zeitstempel (falls zutreffend)

**Service**:
* **name string**: Name des Services beim Drittanbieter (z.B. `transfer_cwa_contact_diary`)
* **operator \*string**: Name des Drittanbieters, der diesen Service betreibt.
* **id \*uuid**: Optional ein eindeutiger Identifier für den Service (vom Anbieter vergeben)
* **properties \*map<string, any>**: Optionale, zusätzliche Eigenschaften, die der Service erfüllen muss (z.B. Kontaktlistenabruf für eine spezifische Ortschaft).

**MessageRequest**:
* **service Service**: Service, an den die Anfrage zu richten ist (ggf. muss zunächst eine Erkennung erfolgen).
* **params map<string, any>**: Methodenparameter (z.B. `tan: 2692`)
* **requested_at \*timestamp**: Zeitstempel der Anfrage (zur Vermeidung von Replay-Angriffen)
* **signature \*Signature**: Signatur mit Zeitstempel (falls zutreffend)

**APIResponse**:
* **status <OK, ERR, ...>**: Statuscode (ggf. feingranularer HTTP-Status)
* **data any**: Anwortdaten
* **error \*Error**: Fehlerinformationen (falls zutreffend)

**Signature**:

* **type string**: Signaturtyp (z.B. ECDSA)
* **data any**: Signaturdaten (typ-spezifisch)

**Error**:

* **code int**: Fehlercode
* **description string**: Fehlerbeschreibung
* **data \*any**: Strukturierte Fehlerdaten (falls zutreffend)

## Server-Komponenten

Es werden Server für zwei Akteure benötigt: Gesundheitsämter (GA) und Anbieter. Beide verfügen über folgende gemeinsame Komponenten:

* Interner Server (JSON-RPC / REST) (`InternalServer`)
* Konfigurationsmanagement (Laden & Validierung von Konfiguration) (`Settings`)
* Nachrichten-Zwischenspeicherung (`MessageStore`)
* Logging (`Logging`)

Der GA-Server besitzt zudem folgende Komponenten:

* Messaging-Dienst zur Verbindung mit dem externen Message-Server (`MessagingClient`)
* Diensterkennung (`ServiceDiscovery`)

Der Anbieter-Server besitzt weiterhin folgende Komponenten:

* Messaging-Server zum Annehmen und Senden von Nachrichten ans GA (`MessagingServer`)

### `MessageStore`

Der `MessageStore` speichert ein- und ausgehende Nachrichten. Er stellt hierbei sicher, dass der Server nicht von Nachrichten überflutet werden kann und limitiert die Anzahl zwischengespeicherter Nachrichten. Jeder Anbieter sowie das GA selbst kann hierbei eine Maximalzahl von Nachrichten speichern. Bei der synchronen Zustellung von Nachrichten kann auf eine Zwischenspeicherung verzichtet werden, jedoch ist diese erforderlich um z.B. Ausfälle externer oder interner Dienste abzufedern.

Insbesondere für den Anbieter-Server ist der `MessageStore` essentiell, da dieser auf eine eingehende Verbindung durch ein GA warten und dementsprechend Nachrichten zwischenspeichern muss.

Um den Verlust von Nachrichten auszuschließen kann ein verschlüsseltes Write-Ahead-Log geführt werden, welches Nachrichten ausfallsicher speichert und diese nach einem Absturz oder Neustart des Servers wieder laden kann. Eine solche Persistierung sollte jedoch im Allgemeinen nach Möglichkeit vermieden werden.

### `MessagingClient`

Der `MessagingClient` stellt Nachrichten an externe Anbieter zu und empfängt Nachrichten von diesen. Er hält hierzu ggf. eine aktive Verbindung zu spezifischen Anbietern offen um von diesen Nachrichten empfangen zu können, oder initiiert bedarfsorientiert Verbindungen um Nachrichten zu senden.

## Abläufe

Die folgenden Abschnitte beschreiben die wesentlichen Programmabläufe der Server.

### GA-Server

Der GA-Server stellt bei der Ausführung einen internen HTTPs-Server bereit und öffnet nach draußen Verbindungen zu relevanten Anbietern.

* Einlesen der Konfigurationseinstellungen
* Validieren der Konfigurationseinstellungen. Bei Fehler Abbruch.
* Öffnen des internen Servers.
* Initialisieren des `MessageStore`.
* Warten auf Anfragen vom internen Server.

#### Annehmen einer Nachricht vom internen Server

Der GA-Server nimmt aus der internen Infrastruktur Nachrichten entgegen und leitet diese an den oder die passenden Anbieter weiter.

* Nachricht entgegennehmen und im `MessageStore` speichern. Weiterleitungsprozess anstoßen. Bei synchroner Verarbeitung Blockierung bis zu einem Time-Out oder erfolgreicher Übermittlung der Nachricht. Bei asynchroner Verarbeitung direkte Rückgabe der `Message` UUID.

#### Zustellen einer Nachricht an den internen Server

Der GA-Server übermittelt von den Anbietern erhaltene Nachrichten aktiv an die interne Infrastuktur.

* Hook-Endpunkt aufrufen und versuchen, Nachricht zu übermitteln. Bei Erfolg Nachricht aus `MessageStore` löschen, ansonsten Fehlerinformation protokollieren.

#### Zustellen einer Nachricht an einen externen Dienst

Der GA-Server stellt Nachrichten an externe Dienste zu.

* Ermitteln, welche(r) Dienst(e) für die Nachricht relevant sind. Falls mehrere Dienste zuständig sind Kopien der Nachricht für jeden Dienst anlegen.
* Jede voll adressierte Nachricht an den `MessageClient` übergeben
* Verbindung zum externen Anbieter herstellen, falls noch keine besteht. Versuchen, alle Nachrichten an den Anbieter zu übertragen.

#### Verbindungs-Management

Der GA-Server hält nach Bedarf Verbindungen zu relevanten Anbietern offen.

* Ermitteln, zu welchen Anbietern eine ständige Verbindung aufgebaut werden soll.
* Zu jedem zutreffenden Anbieter eine Verbindung aufbauen und aufrecht erhalten.
* Nach Bedarf zusätzliche Verbindungen zu Anbietern aufbauen, für die Nachrichten vorliegen.

#### Annehmen von Nachrichten von einem externen Dienst

Der GA-Server nimmt auch Nachrichten von externen Anbietern an. Hierbei wird über Das `ServiceDirectory` ermittelt, welche Nachrichten ein Anbieter an das GA senden darf (und ob z.B. Nachrichten ohne Initiierung durch das GA gesendet werden dürfen).

### Anbieter-Server

Der Anbieter-Server stellt bei der Ausführung sowohl einen internen HTTPs-Server als auch einen externen gRPC-Server bereit.

* Einlesen der Konfigurationseinstellungen
* Validieren der Konfigurationseinstellungen. Bei Fehler Abbruch.
* Öffnen des internen HTTPs-Servers.
* Öffnen des externen gRPC-Servers.
* Initialisieren des `MessageStore`.
* Warten auf Anfragen vom internen Server und externen Server.

Prozesse zum Annehmen und Zustellen von Nachrichten vom internen Server sowie zur Annahme von externen Nachrichten sind im Wesentlichen identisch zum GA-Server. Der einzige relevante Unterschied ist, dass der Anbieter-Server auf eine Verbindungsaufnahme durch den GA-Server warten muss und diese nicht selbst initiieren kann.