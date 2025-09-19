#!/usr/bin/env python3
"""
Sample application for Image Builder demo
This demonstrates building OCI images from directory contexts
"""

import os
import time
from http.server import HTTPServer, SimpleHTTPRequestHandler

class DemoHandler(SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'OK')
        elif self.path == '/':
            self.send_response(200)
            self.send_header('Content-type', 'text/html')
            self.end_headers()
            html = """
            <html>
            <head><title>Image Builder Demo</title></head>
            <body>
                <h1>Image Builder Demo Application</h1>
                <p>Built with OCI Image Builder</p>
                <p>Timestamp: {}</p>
                <p>Environment: {}</p>
            </body>
            </html>
            """.format(time.strftime('%Y-%m-%d %H:%M:%S'), 
                      os.environ.get('ENVIRONMENT', 'demo'))
            self.wfile.write(html.encode())
        else:
            super().do_GET()

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8080))
    server = HTTPServer(('', port), DemoHandler)
    print(f"Starting demo server on port {port}")
    print("Health check: /health")
    server.serve_forever()