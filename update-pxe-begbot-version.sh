#!/bin/bash
cd /home/simon/repos/begbot
git pull origin main  # Använd din branch, t.ex. main
# Installera nya beroenden om det behövs
npm install  # Eller pip install -r requirements.txt
# Starta om dev servern
./dev.nu
