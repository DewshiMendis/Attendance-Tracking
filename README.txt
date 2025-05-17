ATTENDANCE APP
=============

This is a simple attendance tracking application built with Go for the backend and HTML/JavaScript/CSS for the frontend. It uses SQLite for data storage.

Prerequisites
-------------
- Go (version 1.16 or later) must be installed on your system.
- A web browser to access the frontend.
- make sure to allow popus in browser settings otherwise when registering as a new user the qr code will not pop up.
- make sure to install google authenticator in your mobile device.

Setup Instructions
------------------
1. Clone or download the project to your local machine.
2. Navigate to the project directory:
   
   cd ATTENDANCE_APP
  
3. Install the required Go dependencies:
   
   go mod tidy
   
4. Ensure the `attendance.db` SQLite database file is present in the project root.

Running the Application
-----------------------
1. Start the Go server:
   
   go run main.go
   
2. Open a web browser and navigate to:
   
   http://localhost:8080
   
   The application should now be running, and you can interact with it via the browser.(as mentioned previously allow popups in browser.)

   As a new user you will be able to register your self by scanning the qr code and then log in again by entering your password and otp.
   
   After logging in you will be able to mark your attendance.

Project Structure
-----------------
- db/ - Contains database-related Go code (db.go).
- services/ - Contains service logic for authentication (auth_service.go) and OTP (otp_service.go).
- static/ - Contains frontend files (index.html, scripts.js, styles.css).
- utils/ - Contains utility functions (prompt.go).
- attendance.db - SQLite database file.
- main.g - Main application entry point.
- *_qrcode.png - QR code images for users.

Troubleshooting
---------------
- If the server fails to start, ensure all dependencies are installed (`go mod tidy`) and the port 8080 is not in use.
- If the database is missing or corrupted, you may need to recreate it based on the schema defined in `db/db.go`.