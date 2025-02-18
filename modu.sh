#!/usr/bin/env bash

git init;
sleep 1;
git init;
git add -A;
git commit -m "first commit"
git branch -M main
git remote add origin git@github.com:thnkr-one/pdfripper.git
git push -u origin main