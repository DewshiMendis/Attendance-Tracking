let currentDate = new Date();
let attendanceDates = [];
let loggedInUsername = '';

function showSection(sectionId) {
    document.querySelectorAll('.section').forEach(section => {
        section.classList.remove('active');
    });
    document.getElementById(sectionId).classList.add('active');
    clearMessages();
}

function clearMessages() {
    document.querySelectorAll('.error, .success').forEach(el => {
        el.textContent = '';
    });
    document.getElementById('admin-action-form').innerHTML = '';
}

// Calendar functions
function updateCalendar() {
    const months = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];
    const month = currentDate.getMonth();
    const year = currentDate.getFullYear();

    document.getElementById('calendar-month-year').textContent = `${months[month]} ${year}`;

    const firstDay = new Date(year, month, 1).getDay();
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    const today = new Date();
    const isCurrentMonth = today.getMonth() === month && today.getFullYear() === year;
    const todayDate = today.getDate();

    let grid = `
        <div class="day-name">Su</div>
        <div class="day-name">Mo</div>
        <div class="day-name">Tu</div>
        <div class="day-name">We</div>
        <div class="day-name">Th</div>
        <div class="day-name">Fr</div>
        <div class="day-name">Sa</div>
    `;

    for (let i = 0; i < firstDay; i++) {
        grid += `<div class="filler"></div>`;
    }

    for (let day = 1; day <= daysInMonth; day++) {
        const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
        const isAttended = attendanceDates.includes(dateStr);
        const isToday = isCurrentMonth && day === todayDate;
        const classes = `day ${isAttended ? 'attended' : ''} ${isToday ? 'today' : ''}`;
        grid += `<div class="${classes}">${day}</div>`;
    }

    document.getElementById('calendar-grid').innerHTML = grid;
}

function prevMonth() {
    currentDate.setMonth(currentDate.getMonth() - 1);
    updateCalendar();
}

function nextMonth() {
    currentDate.setMonth(currentDate.getMonth() + 1);
    updateCalendar();
}

async function fetchAttendanceDates(username) {
    if (!username) return;
    try {
        const response = await fetch(`/api/attendance/dates?username=${encodeURIComponent(username)}`);
        const data = await response.json();
        if (response.ok) {
            attendanceDates = data.dates || [];
            updateCalendar();
        }
    } catch (err) {
        console.error('Error fetching attendance dates:', err);
    }
}

// Initialize calendar
document.addEventListener('DOMContentLoaded', () => {
    updateCalendar();
});

// Show OTP popup
function showOTPPopup(username, secret, qrCodeUrl) {
    const popup = document.createElement('div');
    popup.className = 'otp-popup';
    popup.innerHTML = `
        <div class="otp-popup-content">
            <h3>Scan QR Code and Enter OTP</h3>
            <img src="${qrCodeUrl}" alt="QR Code">
            <form id="otp-form">
                <input type="text" id="otp-input" placeholder="Enter OTP" required>
                <button type="submit">Verify OTP</button>
                <button type="button" class="close-button">Close</button>
            </form>
            <div id="otp-error" class="error"></div>
            <div id="otp-success" class="success"></div>
        </div>
    `;
    document.body.appendChild(popup);

    // Handle OTP submission
    document.getElementById('otp-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const otp = document.getElementById('otp-input').value;
        const errorEl = document.getElementById('otp-error');
        const successEl = document.getElementById('otp-success');

        try {
            const response = await fetch('/api/verify-otp', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, otp, secret })
            });
            const data = await response.json();
            if (!response.ok) {
                errorEl.textContent = data.message || 'OTP verification failed';
                return;
            }
            successEl.textContent = 'OTP verified successfully! Please log in.';
            setTimeout(() => {
                popup.remove();
                showSection('login-section');
            }, 2000);
        } catch (err) {
            errorEl.textContent = 'Error: Could not connect to server';
        }
    });

    // Handle close button
    popup.querySelector('.close-button').addEventListener('click', () => {
        popup.remove();
    });
}

// Handle registration
document.getElementById('register-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const username = document.getElementById('reg-username').value;
    const password = document.getElementById('reg-password').value;
    const errorEl = document.getElementById('register-error');
    const successEl = document.getElementById('register-success');

    try {
        const response = await fetch('/api/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        const data = await response.json();
        if (!response.ok) {
            errorEl.textContent = data.message || 'Registration failed';
            return;
        }
        successEl.textContent = 'Registered successfully! Opening OTP verification...';
        setTimeout(() => {
            showOTPPopup(username, data.secret, data.qrCodeUrl);
        }, 1000);
    } catch (err) {
        errorEl.textContent = 'Error: Could not connect to server';
    }
});

// Handle login
document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;
    const otp = document.getElementById('login-otp').value;
    const errorEl = document.getElementById('login-error');

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password, otp })
        });
        const data = await response.json();
        if (!response.ok) {
            errorEl.textContent = data.message || 'Login failed';
            return;
        }
        loggedInUsername = username;
        fetchAttendanceDates(username);
        if (data.role === 'admin') {
            showSection('admin-section');
        } else {
            showSection('user-section');
        }
    } catch (err) {
        errorEl.textContent = 'Error: Could not connect to server';
    }
});

// Handle attendance recording
async function recordAttendance() {
    const messageEl = document.getElementById('user-message');
    if (!loggedInUsername) {
        messageEl.textContent = 'Please log in again';
        messageEl.className = 'error';
        return;
    }

    try {
        const response = await fetch('/api/attendance', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: loggedInUsername })
        });
        const data = await response.json();
        if (!response.ok) {
            messageEl.textContent = data.message || 'Failed to record attendance';
            messageEl.className = 'error';
            return;
        }
        messageEl.textContent = data.message;
        messageEl.className = 'success';
        fetchAttendanceDates(loggedInUsername); // Refresh calendar
    } catch (err) {
        messageEl.textContent = 'Error: Could not connect to server';
        messageEl.className = 'error';
    }
}

// Handle admin actions
function showAdminAction(action) {
    const formContainer = document.getElementById('admin-action-form');
    formContainer.innerHTML = '';
    const messageEl = document.getElementById('admin-message');

    if (action === 'reset-password') {
        formContainer.innerHTML = `
            <form id="reset-password-form">
                <input type="text" id="admin-username" placeholder="Admin Username" required>
                <input type="text" id="target-username" placeholder="Target Username" required>
                <input type="password" id="new-password" placeholder="New Password" required>
                <button type="submit">Reset Password</button>
            </form>
        `;
        document.getElementById('reset-password-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const adminUsername = document.getElementById('admin-username').value;
            const targetUsername = document.getElementById('target-username').value;
            const newPassword = document.getElementById('new-password').value;

            try {
                const response = await fetch('/api/admin/reset-password', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ adminUsername, targetUsername, newPassword })
                });
                const data = await response.json();
                messageEl.textContent = data.message;
                messageEl.className = response.ok ? 'success' : 'error';
            } catch (err) {
                messageEl.textContent = 'Error: Could not connect to server';
                messageEl.className = 'error';
            }
        });
    } else if (action === 'change-role') {
        formContainer.innerHTML = `
            <form id="change-role-form">
                <input type="text" id="admin-username" placeholder="Admin Username" required>
                <input type="text" id="target-username" placeholder="Target Username" required>
                <input type="text" id="new-role" placeholder="New Role (user/admin)" required>
                <button type="submit">Change Role</button>
            </form>
        `;
        document.getElementById('change-role-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const adminUsername = document.getElementById('admin-username').value;
            const targetUsername = document.getElementById('target-username').value;
            const newRole = document.getElementById('new-role').value;

            try {
                const response = await fetch('/api/admin/change-role', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ adminUsername, targetUsername, newRole })
                });
                const data = await response.json();
                messageEl.textContent = data.message;
                messageEl.className = response.ok ? 'success' : 'error';
            } catch (err) {
                messageEl.textContent = 'Error: Could not connect to server';
                messageEl.className = 'error';
            }
        });
    } else if (action === 'delete-user') {
        formContainer.innerHTML = `
            <form id="delete-user-form">
                <input type="text" id="admin-username" placeholder="Admin Username" required>
                <input type="text" id="target-username" placeholder="Target Username" required>
                <button type="submit">Delete User</button>
            </form>
        `;
        document.getElementById('delete-user-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const adminUsername = document.getElementById('admin-username').value;
            const targetUsername = document.getElementById('target-username').value;

            try {
                const response = await fetch('/api/admin/delete-user', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ adminUsername, targetUsername })
                });
                const data = await response.json();
                messageEl.textContent = data.message;
                messageEl.className = response.ok ? 'success' : 'error';
            } catch (err) {
                messageEl.textContent = 'Error: Could not connect to server';
                messageEl.className = 'error';
            }
        });
    } else if (action === 'list-users') {
        fetch('/api/admin/list-users')
            .then(response => response.json())
            .then(data => {
                formContainer.innerHTML = `<ul>${data.users.map(user => `<li>${user}</li>`).join('')}</ul>`;
                messageEl.textContent = 'Users listed successfully';
                messageEl.className = 'success';
            })
            .catch(err => {
                messageEl.textContent = 'Error: Could not connect to server';
                messageEl.className = 'error';
            });
    }
}