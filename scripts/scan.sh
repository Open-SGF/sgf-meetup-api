#!/usr/bin/env zsh
aws dynamodb scan --table-name items --endpoint http://localhost:8000
