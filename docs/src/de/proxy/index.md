# Der EPS TLS-Passthrough-Proxy

Der EPS-Server bietet einen TLS-Passthrough-Proxy-Dienst, der eingehende TLS-Verbindungen von öffentlichen Clients an einen internen Server weiterleiten kann, ohne die TLS-Verbindung zu beenden. Der Dienst besteht aus zwei Servern:

* Ein **öffentlicher Proxy**, der auf einem öffentlich zugänglichen TCP-Port auf eingehende TLS-Verbindungen lauscht und auf einem anderen TCP-Port auf Verbindungen von privaten Proxy-Servern.
* Ein **privater Proxy**, der auf keinem TCP-Port lauscht und sich stattdessen aktiv mit dem öffentlichen Proxy-Server verbindet, wenn eine Verbindung für ihn verfügbar ist. Er leitet diese Verbindung (wiederum ohne TLS zu beenden) an einen internen Server weiter, der sie dann bearbeitet.

Sowohl die privaten als auch die öffentlichen Proxys haben zugehörige EPS-Server, über die sie miteinander kommunizieren. 

Der private Proxy kann eingehende Verbindungen an den öffentlichen Proxy **ankündigen**. Immer wenn eine neue Verbindung diesen Proxy erreicht, vergleicht er sie mit den Ankündigungen, die er erhalten hat. Wenn eine Übereinstimmung gefunden wird, benachrichtigt er den privaten Proxy über eine eingehende Verbindung über das EPS-Netzwerk und sendet ihm außerdem ein zufälliges Token. Der private Proxy öffnet eine Verbindung zum öffentlichen Proxy, sendet das Token und übernimmt den TCP-Stream. Er leitet ihn an einen internen Server weiter.

## Demo

Um diesen Mechanismus zu demonstrieren, haben wir eine Beispielkonfiguration vorbereitet. Führen Sie einfach die folgenden Schnipsel in verschiedenen Terminals aus (aus dem Hauptverzeichnis im Repository):

```bash
# prepare the binaries
make && make examples
# first terminal
internal-server #will open a JSON-RPC server on port 8888
# second terminal (public proxy)
PROXY_SETTINGS=settings/dev/roles/public-proxy-1 proxy run public
# third terminal (private proxy)
PROXY_SETTINGS=settings/dev/roles/private-proxy-1 proxy run private
# fourth terminal (public proxy EPS server)
EPS_SETTINGS=settings/dev/roles/public-proxy-eps-1 eps server run
# fifth terminal (private proxy EPS server)
EPS_SETTINGS=settings/dev/roles/private-proxy-eps-1 eps server run
```

Wenn alle Dienste laufen, sollten Sie in der Lage sein, eine Anfrage an den Proxy zu senden über

```bash
curl --cacert settings/dev/certs/root.crt --resolve test.internal-server.local:4433:127.0.0.1 https://test.internal-server.local:4433/jsonrpc | jq .

```

Dies sollte die folgenden JSON-Daten zurückgeben:

```json
{
  "message": "success"
}
```

Die Anfrage, die Sie gesendet haben, erreichte den lokalen TLS-Server auf Port 8888 über die beiden Proxys, die über das EPS-Netzwerk kommunizierten, um die Verbindung zu vermitteln. Toll, nicht wahr?

## Stresstest

Sie können den Server auch einem Stresstest mit parallelen Anfragen unterziehen, indem Sie den `parallel`
verwenden:

```bash
eq 1 2000000 | parallel -j 25 curl --cacert settings/dev/certs/root.crt --resolve test.internal-server.local:4433:127.0.0.1 https://test.internal-server.local:4433/jsonrpc --data "{}"
```

Dadurch wird versucht, 25 Anfragen parallel an den Server zu senden.
