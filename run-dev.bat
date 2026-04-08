@echo off
echo ========================================================
echo 🚀 MEMBANGUNKAN INFRASTRUKTUR TIER 1 (IDE DEV MODE) 🚀
echo ========================================================

echo.
echo [1/2] Menenggelamkan diri ke Docker: Membangun Database Master, Replica, dan Redis...
docker-compose up -d postgres-master postgres-replica redis-cache

echo.
echo [2/2] Menyalakan Traktor API Golang Fiber V2 (Port 9000)...
go run ./backend-api

pause
