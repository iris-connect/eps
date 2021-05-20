# Erste Schritte

Die Integration mit der IRIS-Infrastruktur unter Verwendung des EPS-Servers ist einfach (hoffen wir). Zunächst müssen Sie den `eps` Server zusammen mit den Einstellungen und Zertifikaten, die wir Ihnen zur Verfügung gestellt haben, bereitstellen. Dies ist so einfach wie das Herunterladen der neuesten Version von `eps` von unserem Server, das Entpacken des Einstellungsarchivs, das wir Ihnen zur Verfügung gestellt haben, und das Ausführen von

`` `bash
EPS_SETTINGS=Pfad/zu/Einstellungen eps Serverlauf
`` `

Dies sollte einen lokalen JSON-RPC-Server auf Port `5555` öffnen, mit dem Sie sich über TLS verbinden können (dazu müssen Sie das Root-CA-Zertifikat zu Ihrer Zertifikatskette hinzufügen). Dieser Server ist Ihr Gateway zu allen IRIS-Diensten. Schauen Sie einfach nach den Diensten, die ein bestimmter Betreiber anbietet, und senden Sie eine Anfrage, die den Namen des Betreibers und die Dienstmethode enthält, die Sie aufrufen möchten. Um z. B. mit dem Dienst "locations" zu interagieren, der vom Betreiber "ls-1" bereitgestellt wird, würden Sie einfach eine JSON-RPC-Nachricht wie diese senden:

`` `json
{
	"Methode": "ls-1.add",
	"id": "1",
	"Params": {
		"Name": "Ginos",
		"id": "af5ca4da5caa"
	},
	"jsonrpc": "2.0"
}
`` `

Das Gateway kümmert sich darum, diese Nachricht an den richtigen Dienst weiterzuleiten und eine Antwort an Sie zurückzuschicken.

Wenn Sie Anfragen von anderen Diensten im IRIS-Ökosystem akzeptieren möchten, können Sie die `jsonrpc_client` verwenden. Dabei geben Sie einfach einen API-Endpunkt an, an den eingehende Anfragen mit der gleichen Syntax wie oben zugestellt werden sollen.

Das war's!

## Asynchrone Anrufe

Die Anrufe, die wir oben gesehen haben, waren alle synchron, d. h. ein Anruf führte zu einer direkten Antwort. Manchmal müssen Aufrufe jedoch asynchron sein, z. B. weil die Beantwortung Zeit benötigt. Wenn Sie einen asynchronen Aufruf an einen anderen Dienst tätigen, erhalten Sie zunächst eine Bestätigung zurück. Sobald der von Ihnen aufgerufene Dienst eine Antwort bereit hat, sendet er diese über das Netzwerk `eps` an Sie zurück, wobei er dieselbe `id` verwendet, die Sie angegeben haben (wodurch Sie die Antwort mit Ihrer Anfrage abgleichen können). Ebenso können Sie auf Aufrufe von anderen Diensten asynchron reagieren, indem Sie die Antwort einfach mit dem Methodennamen `respond` (ohne Dienstnamen) an Ihren lokalen JSON-RPC-Server schieben. Vergessen Sie nicht, dieselbe `id` anzugeben, die Sie mit der ursprünglichen Anfrage erhalten haben, da diese die "Rücksendeadresse" der Anfrage enthält.

## Integration Beispiel

Um eine konkrete Vorstellung von der Integration mit der IRIS-Infrastruktur unter Verwendung des EPS-Servers zu bekommen, haben wir ein einfaches Demo-Setup erstellt, das alle Komponenten veranschaulicht. Die Demo besteht aus drei Komponenten:

* Ein `eps` Server, der einen `health department` simuliert, namens `hd-1`
* Eine `eps` Server-Simulation eines Betreibers, der einen "Standort"-Dienst anbietet, namens `ls-1`
* Der tatsächliche vom Betreiber angebotene Ortungsdienst `eps-ls` `ls-1`

## Aufstehen und loslegen

Lesen Sie bitte zuerst in der README nach, wie Sie alle notwendigen TLS-Zertifikate erstellen und die Software bauen. Starten Sie dann die einzelnen Dienste auf verschiedenen Terminals:

`` `bash
# führen Sie den `eps` Server des "Standorte"-Operators ls-1 aus
EPS_SETTINGS=settings/dev/roles/ls-1 eps --level debug server run
# Betreiben Sie den `eps` Server des Gesundheitsamtes hd-1
EPS_SETTINGS=settings/dev/roles/hd-1 eps --level debug server run
# Führen Sie den Dienst "Standorte" aus
eps-ls
`` `

## Testen

Jetzt sollte Ihr System einsatzbereit sein. Der Demo-Dienst "locations" bietet eine einfache, authentifizierungsfreie JSON-RPC-Schnittstelle mit zwei Methoden: `add`, die einen Ort zur Datenbank hinzufügt, und `lookup`, die einen Ort anhand seiner `name` nachschlägt. Zum Beispiel, um dem Dienst einen Ort hinzuzufügen:

`` `bash
curl --cacert settings/dev/certs/root.crt --resolve hd-1:5555:127.0.0.1 https://hd-1:5555/jsonrpc --header "Content-Type: application/json" --data '{"method": "ls-1.add", "id": "1", "params": {"name": "Ginos", "id": "af5ca4da5caa"}, "jsonrpc": "2.0"}' 2&gt;/dev/null | jq 
`` `

Dies sollte eine einfache JSON-Antwort zurückgeben:

`` `json
{
  "jsonrpc": "2.0",
  "Ergebnis": {
    "_": "ok"
  },
  "id": "1"
}
`` `

Die Anfrage ging zunächst an den `eps` Server des Gesundheitsamtes, wurde über gRPC zunächst an den `ls-1` 's `eps` Server weitergeleitet und wurde dann an die JSON-RPC API des dort laufenden lokalen `eps-ls` Dienstes übergeben. Das Ergebnis wurde dann entlang der gesamten Kette zurückgereicht.

Sie können auch eine Suche nach dem Ort durchführen, den Sie gerade hinzugefügt haben:

`` `bash
curl --cacert settings/dev/certs/root.crt --resolve hd-1:5555:127.0.0.1 https://hd-1:5555/jsonrpc --header "Content-Type: application/json" --data '{"method": "ls-1.lookup", "id": "1", "params": {"name": "Ginos"}, "jsonrpc": "2.0"}' 2&gt;/dev/null | jq .
`` `

die Folgendes zurückgeben sollte

`` `json
{
  "jsonrpc": "2.0",
  "Ergebnis": {
    "id": "af5ca4da5caa",
    "Name": "Ginos"
  },
  "id": "1"
}
`` `

Daher ist die Interaktion mit dem entfernten "locations"-Dienst genauso wie der Aufruf eines lokalen JSON-RPC-Dienstes, außer dass Sie den Namen des Operators, der den Dienst ausführt, `ls-1.lookup`, angeben, anstatt einfach `lookup` aufzurufen.

