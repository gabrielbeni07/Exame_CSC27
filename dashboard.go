package main

import (
	"net/http"
	"strings"
	"sync"
)

var mutex sync.RWMutex

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	html := `
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Air Traffic Dashboard</title>
        <style>
            body { font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 0; }
            .container { max-width: 800px; margin: 20px auto; padding: 20px; background: white; box-shadow: 0 0 10px rgba(0, 0, 0, 0.1); }
            h1 { text-align: center; color: #333; }
            table { width: 100%; border-collapse: collapse; margin-top: 20px; }
            table th, table td { border: 1px solid #ddd; padding: 8px; text-align: left; }
            table th { background: #f8f8f8; color: #333; }
            .landing { background: #d1ecf1; color: #0c5460; }
            .takeoff { background: #f8d7da; color: #721c24; }
            .updates { background: #fff3cd; color: #856404; }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>Air Traffic Dashboard</h1>

            <h2>Queue (Pending Requests)</h2>
            <table>
                <thead>
                    <tr>
                        <th>Aircraft ID</th>
                        <th>Type</th>
                        <th>Timestamp</th>
                    </tr>
                </thead>
                <tbody>
                    {{.QueueRows}}
                </tbody>
            </table>

            <h2>Updates (Real-Time Notifications)</h2>
            <table>
                <thead>
                    <tr>
                        <th>Type</th>
                        <th>Details</th>
                        <th>Timestamp</th>
                    </tr>
                </thead>
                <tbody>
                    {{.UpdateRows}}
                </tbody>
            </table>
        </div>
    </body>
    </html>
    `

	mutex.RLock()
	defer mutex.RUnlock()

	var queueRows string
	for _, msg := range queueState {
		payload := msg.Payload.(map[string]string)
		rowClass := payload["type"]
		queueRows += `<tr class="` + rowClass + `"><td>` + payload["aircraft_id"] + `</td><td>` + payload["type"] + `</td><td>` + payload["timestamp"] + `</td></tr>`
	}

	var updateRows string
	for _, msg := range updatesState {
		payload := msg.Payload.(map[string]string)
		updateRows += `<tr class="updates"><td>` + payload["update_type"] + `</td><td>` + payload["details"] + `</td><td>` + payload["timestamp"] + `</td></tr>`
	}

	html = strings.Replace(html, "{{.QueueRows}}", queueRows, 1)
	html = strings.Replace(html, "{{.UpdateRows}}", updateRows, 1)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
