db-init:
	rm -rf  ./database/database.db
	sqlite3 ./database/database.db < ./database/dump.sql