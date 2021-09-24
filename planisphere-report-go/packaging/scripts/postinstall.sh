#!/bin/bash

systemctl daemon-reload
systemctl enable --now planisphere-report.timer
