// 포털 애플리케이션 메인 로직
let currentSession = null;

// DOM 요소들
const loginBtn = document.getElementById('loginBtn');
const consoleBtn = document.getElementById('consoleBtn');
const logoutBtn = document.getElementById('logoutBtn');
const userInfo = document.getElementById('user-info');
const userName = document.getElementById('user-name');
const loading = document.getElementById('loading');
const error = document.getElementById('error');
const success = document.getElementById('success');

// 페이지 요소들
const loginPage = document.getElementById('login-page');
const dashboardPage = document.getElementById('dashboard-page');

// 페이지 로드 시 사용자 상태 확인
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
});

// 앱 초기화 로직
async function initializeApp() {
    const isLoggedIn = await checkUserStatus();
    const path = window.location.pathname;

    if (isLoggedIn) {
        if (path !== '/dashboard') {
            window.history.pushState({}, '', '/dashboard');
        }
        showDashboardPage(true);
    } else {
        if (path !== '/') {
           window.history.pushState({}, '', '/');
        }
        showLoginPage();
    }
}

async function checkUserStatus() {
    try {
        const response = await fetch('/api/user', { credentials: 'include' });
        if (response.ok) {
            const userData = await response.json();
            window.currentUser = userData.data; 
            return true;
        }
        return false;
    } catch (error) {
        console.error('Error checking user status:', error);
        return false;
    }
}

function showDashboardPage(isLoggedIn) {
    loginPage.style.display = 'none';
    dashboardPage.style.display = 'block';
    if (isLoggedIn && window.currentUser) {
        userName.textContent = `Welcome, ${window.currentUser.user_id}!`;
        userInfo.classList.add('show');
    }
}

function showLoginPage() {
    loginPage.style.display = 'block';
    dashboardPage.style.display = 'none';
}

// 로그인
function login() {
    window.location.href = '/api/login';
}

// 로그아웃
async function logout() {
    try {
        const response = await fetch('/api/logout', {
            method: 'POST',
        });

        if (response.ok) {
            const data = await response.json();
            if (data.logout_url) {
                window.location.href = data.logout_url;
            } else {
                window.location.href = '/';
            }
        } else {
            console.error('Server logout failed');
            window.location.href = '/';
        }
    } catch (error) {
        console.error('Logout request error:', error);
        window.location.href = '/';
    }
}

function showUserInfo(userId) {
    const currentPath = window.location.pathname;
    
    if (currentPath === '/dashboard') {
        userName.textContent = `Welcome, ${userId}!`;
        userInfo.classList.add('show');
    } else {
        window.location.href = '/dashboard';
    }
}

function showLoginSection() {
    const currentPath = window.location.pathname;
    
    if (currentPath === '/dashboard') {
        window.location.href = '/';
    } else {
        loginBtn.style.display = 'inline-block';
        userInfo.classList.remove('show');
        loading.classList.remove('show');
        error.classList.remove('show');
        success.classList.remove('show');
    }
}

async function launchConsole() {
    showLoading();

    try {
        const response = await fetch('/api/launch-console', {
            credentials: 'include'
        });
        
        if (response.ok) {
            const data = await response.json();
            hideLoading();
            showSuccess('Web terminal started successfully!');
            
            setTimeout(() => {
                window.open(data.url, '_blank');
            }, 1000);
        } else {
            const errorData = await response.json();
            hideLoading();
            showError(`Failed to start web terminal: ${errorData.error}`);
        }
    } catch (error) {
        hideLoading();
        showError(`Network error: ${error.message}`);
    }
}

function showLoading() {
    loading.classList.add('show');
    error.classList.remove('show');
    success.classList.remove('show');
}

function hideLoading() {
    loading.classList.remove('show');
}

function showError(message) {
    error.textContent = message;
    error.classList.add('show');
}

function showSuccess(message) {
    success.textContent = message;
    success.classList.add('show');
}

function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}

// 이벤트 리스너 등록
loginBtn.addEventListener('click', login);
consoleBtn.addEventListener('click', launchConsole);
logoutBtn.addEventListener('click', logout);

// 브라우저 뒤로가기/앞으로가기 처리
window.addEventListener('popstate', function() {
    initializeApp();
});
