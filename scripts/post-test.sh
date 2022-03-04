#! /usr/local/bin/bash

curl -X POST -H "Content-Type: application/json" \
-d '{"plan": "9", "payment_method": "linuxize", "email": "linuxize@example.com"}' \
http://localhost:4001/api/create-customer-and-subscribe-to-plan