// 포털 애플리케이션 메인 로직 (SPA 라우팅)

// DOM 요소들
const loginBtn = document.getElementById('loginBtn');
const loginPage = document.getElementById('login-page');
const dashboardPage = document.getElementById('dashboard-page');

// 페이지 로드 시 앱 초기화
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

// 사용자 상태 확인
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

// 대시보드 페이지 표시
function showDashboardPage(isLoggedIn) {
    if (loginPage) loginPage.style.display = 'none';
    if (dashboardPage) dashboardPage.style.display = 'flex';
    
    // dashboard.js의 초기화 함수 호출
    if (window.initializeDashboard) {
        window.initializeDashboard();
    }
}

// 로그인 페이지 표시
function showLoginPage() {
    if (loginPage) loginPage.style.display = 'flex';
    if (dashboardPage) dashboardPage.style.display = 'none';
}

// 로그인
function login() {
    window.location.href = '/api/login';
}

// 이벤트 리스너 등록
if (loginBtn) {
    loginBtn.addEventListener('click', login);
}

// 브라우저 뒤로가기/앞으로가기 처리
window.addEventListener('popstate', function() {
    initializeApp();
});


