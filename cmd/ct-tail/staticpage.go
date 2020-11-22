package main

var staticPage = `<html>
<head>
	<script>
        function printLine(line) {
            if (line == "") {
                return
            }
            today = new Date()
            date = today.toLocaleString()
            line = date + "\t" + line

            let elem = document.getElementById('log')
            let lineElem = document.createElement('pre');
            lineElem.append(line);
            elem.append(lineElem);

            lineElem.scrollIntoView()
        }

        function showEntry(entry) {
            let line = ""
            for (let name of entry.identifiers.dns_names) {
                line += name
            }
            printLine(line)
        }

		window.onload = function runLog() {
            printLine("CT Log tailing started. DNS names for new log entries will appear below.")

            var url = "log"
            fetch(url).then(function (response) {
                let reader = response.body.getReader();
                let decoder = new TextDecoder();
                return readData();
                function readData() {
                    return reader.read().then(function ({value, done}) {
                        let newData = decoder.decode(value, {stream: !done});
                        jsons = newData.split("\n")
                        for (let json of jsons) {
                            if (json != "") {
                                entry = JSON.parse(json)
                                showEntry(entry)
                            }
                        }
                        if (done) {
                            printLine("Stream ended.")
                            return;
                        }
                        return readData();
                    });
                }
            });
        }
	</script>
</head>
<body>
    <div id="log">
    </div>
</body>
</html>
`
