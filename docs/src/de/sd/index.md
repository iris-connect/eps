# Service-Verzeichnis

Das Dienstverzeichnis ist eine zentrale Datenbank, die Informationen über alle Betreiber im IRIS-Ökosystem enthält. Es enthält Informationen darüber, wie die Betreiber erreicht werden können und welche Dienste sie anbieten.

Mit Hilfe des Verzeichnisses können EPS-Server feststellen, ob und wie sie sich mit einem anderen Betreiber verbinden können. Betreiber, die nur über ausgehende Verbindungen verfügen (z. B. `ga-leipzig` im obigen Beispiel), können das Verzeichnis verwenden, um zu erfahren, dass sie möglicherweise asynchrone Antworten von anderen Betreibern (z. B. `ls-1`) erhalten und dann ausgehende Verbindungen zu diesen Betreibern öffnen, über die sie Antworten erhalten können. EPS-Server können das Dienstverzeichnis auch verwenden, um festzustellen, ob sie eine Nachricht von einem bestimmten Betreiber annehmen sollen.

Das Dienstverzeichnis implementiert einen gruppenbasierten Berechtigungsmechanismus. Derzeit existieren nur `yes/no` Berechtigungen (d. h. ein Mitglied einer bestimmten Gruppe kann eine bestimmte Dienstmethode entweder aufrufen oder nicht). Feinkörnigere Berechtigungen (z. B. kann ein Anbieter von Kontaktverfolgungen nur seine eigenen Einträge im Dienst "Standorte" bearbeiten) müssen von den Diensten selbst implementiert werden. Zu diesem Zweck stellt der EPS-Server den Diensten Informationen über die aufrufende Gegenstelle über einen speziellen Parameter (`_caller`) zur Verfügung, der zusammen mit den anderen RPC-Methodenparametern übergeben wird. Diese Struktur enthält auch den aktuellen Eintrag des Aufrufers aus dem Dienstverzeichnis, was es dem aufgerufenen Dienst erleichtert, den Aufrufer zu identifizieren und zu autorisieren.

## Dienstverzeichnis-API

Das EPS-Serverpaket bietet auch einen `sd` API-Server-Befehl, der einen JSON-RPC-Server öffnet, der das Dienstverzeichnis verteilt.

`` `bash
SD_SETTINGS=settings/dev/roles/sd-1 sd run
`` `

Standardmäßig werden damit Änderungsdatensätze in einer Datei gespeichert und abgerufen, die sich unter `/tmp/service-directory.records` befindet. Um das Dienstverzeichnis zurückzusetzen, löschen Sie einfach diese Datei.

## Signatur-Schema

Alle Änderungen im Dienstverzeichnis werden kryptografisch signiert. Dazu besitzt jeder Akteur im EPS-System ein Paar ECDSA-Schlüssel und ein dazugehöriges Zertifikat. Das Serviceverzeichnis ist aus einer Reihe von **Änderungsdatensätzen** aufgebaut. Jeder Änderungssatz enthält den Namen eines **Akteurs**, einen **Abschnitt** und die eigentlichen Daten, die geändert werden sollen.

### Einreichen von Änderungsdatensätzen

Änderungsdatensätze können über die JSON-RPC-API an das Dienstverzeichnis übermittelt werden. Die `eps` CLI bietet dafür eine Funktion über die `sd submit-records` :

`` `bash
EPS_SETTINGS=settings/dev/roles/hd-1 eps sd submit-records settings/dev/directory/001_base.json
`` `

Sie können das Dienstverzeichnis auch zurücksetzen, indem Sie das Flag `--reset` angeben:

`` `
EPS_SETTINGS=settings/dev/roles/hd-1 eps sd submit-records --reset settings/dev/directory/001_base.json
`` `

**Warnung:** Dadurch werden alle vorherigen Datensätze aus dem Dienstverzeichnis gelöscht. Nur Bediener mit einer `sd-admin` Rolle können dies tun.

### Abrufen von Einträgen und Datensätzen

Um Änderungsdatensätze und -einträge von der Dienstverzeichnis-API abzurufen, können Sie die RPC-Aufrufe `getRecords(since)`, `getEntries()` und `getEntry(name)` verwenden, z. B. wie folgt:

`` `bash
curl --key settings/dev/certs/hd-1.key --cert settings/dev/certs/hd-1.crt --cacert settings/dev/certs/root.crt --resolve sd-1:3322:127.0.0.1 https://sd-1:3322/jsonrpc --header "Content-Type: application/json" --data '{"jsonrpc": "2.0", "method": "getRecords", "params": {"since": 0}}'
`` `

### Signierdaten

Das Tool `sdh` enthält einen Befehl `sign`, mit dem wir beliebige JSON-Daten signieren können. Er verwendet die Signaturen, die mit dem Befehl `make certs` Make erzeugt werden. Um zum Beispiel eine JSON-Datei zu signieren, verwenden Sie einfach

`` `
# definieren die SD-Einstellungen
export SD_SETTINGS=settings/dev/roles/private-proxy-1/sdh

# einen Dienstverzeichniseintrag signieren
sdh sign settings/dev/roles/private-proxy-1/sdh/entry.json
`` `

Die Ausgabe sollte z. B. so aussehen:

`` `json
{
  "Signatur": {
    "r": "67488385997031737348502334621054744305438368369525250023542608571625588981387",
    "s": "110557266089828975725234959115295121652814407881082688883738138814924173982570",
    "c": "-----BEGIN CERTIFICATE-----\nMIIC1TCCAb2gAwIBAgIUe3+081Bi4Z0DXDdeBhfZZOAs4OwwDQYJKoZIhvcNAQEL\nBQAwaTELMAkGA1UEBhMCREUxDzANBgNVBAgMBkJlcmxpbjEPMA0GA1UEBwwGQmVy\nbGluMQ0wCwYDVQQKDARJUklTMQswCQYDVQQLDAJJVDEcMBoGA1UEAwwTVGVzdGlu\nZy1EZXZlbG9wbWVudDAeFw0yMTA1MTExMTMzNDBaFw0yMjA5MjMxMTMzNDBaMGUx\nCzAJBgNVBAYTAkRFMQ8wDQYDVQQIDAZCZXJsaW4xDzANBgNVBAcMBkJlcmxpbjEN\nMAsGA1UECgwESVJJUzELMAkGA1UECwwCSVQxGDAWBgNVBAMMD3ByaXZhdGUtcHJv\neHktMTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABLHlILI5POvEDJc96W0dbag7\nFt8BVmitGqwS5jarYRwOUe/PiQ8tMBkMw9X/2U8G1qGYQb/CiRDh1DDy/Eh/mGKj\nRDBCMDMGA1UdEQQsMCqCD3ByaXZhdGUtcHJveHktMYIXKi5wcml2YXRlLXByb3h5\nLTEubG9jYWwwCwYDVR0PBAQDAgeAMA0GCSqGSIb3DQEBCwUAA4IBAQAmUESzD1ls\nmpECtRlinhiUduif9nVddtLeW/Ui86PHkS50vjSOVHY7ZHrfWbFB4/p4bwm8Sp1/\npFHx4WyuHiow5Ah3HV9afDcgyWBd1V8ijIFOlNF27u/caVsa9gV7iDVJ+6mBXKkf\nCgNI2bA2WoOVXQMwRoow4vSYrVAdM/Eyq8PHYOHkGqdd4uASG5df4vE+gnB2z9WD\nFuxkVYkncVP5OB+N7EAkQrVjrITdiSN0yYAVWFKz1IEnPF7GRW6KsPHW9lJeePeD\n1gLNh2KF6drrXT2PIIYVB31uepSoCqFnUUDcC/PX0qHu8jilvr/pTzhFUWbuX+Ja\nfaIRxqWB0frZ\n-----END
ZERTIFIKAT-----\n"
  },
  "Daten": {
    "foo": "bar",
    "Name": "Privat-Proxy-1"
  }
}
`` `

Bevor wir solche Daten importieren, können wir die Signatur mit dem Befehl `verify` überprüfen (Sie müssen die erwartete `name` des Unterzeichners angeben):

`` `bash
export SD_SETTINGS=settings/dev/roles/sd-1/sdh
sdh verify signed.json private-proxy-1
`` `

Wenn die Signatur gültig ist, lautet der Exit-Code `0`, andernfalls `1`.
