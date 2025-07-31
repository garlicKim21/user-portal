// 포털 애플리케이션 메인 로직

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

// 페이지 로드 시 사용자 상태 확인
document.addEventListener('DOMContentLoaded', function() {
    checkUserStatus();
});

// URL에서 세션 파라미터 확인
function getSessionFromURL() {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get('session');
}

// 사용자 상태 확인
async function checkUserStatus() {
    const session = getSessionFromURL();
    if (session) {
        currentSession = session;
        try {
            const response = await fetch(`/api/user?session=${session}`);
            if (response.ok) {
                const userData = await response.json();
                showUserInfo(userData.user_id);
            } else {
                showLoginSection();
            }
        } catch (error) {
            console.error('Error checking user status:', error);
            showLoginSection();
        }
    } else {
        showLoginSection();
    }
}

// 로그인
function login() {
    window.location.href = '/api/login';
}

// 로그아웃
function logout() {
    currentSession = null;
    showLoginSection();
    // URL에서 세션 파라미터 제거
    window.history.replaceState({}, document.title, window.location.pathname);
}

// 사용자 정보 표시
function showUserInfo(userId) {
    loginBtn.style.display = 'none';
    userName.textContent = `Welcome, ${userId}!`;
    userInfo.classList.add('show');
}

// 로그인 섹션 표시
function showLoginSection() {
    loginBtn.style.display = 'inline-block';
    userInfo.classList.remove('show');
    loading.classList.remove('show');
    error.classList.remove('show');
    success.classList.remove('show');
}

// 웹 콘솔 실행
async function launchConsole() {
    if (!currentSession) {
        showError('No session found. Please login again.');
        return;
    }

    showLoading();

    try {
        const response = await fetch(`/api/launch-console?session=${currentSession}`);
        
        if (response.ok) {
            const data = await response.json();
            hideLoading();
            showSuccess('Web terminal started successfully!');
            
            // 새 탭에서 웹 콘솔 열기
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

// 로딩 표시
function showLoading() {
    loading.classList.add('show');
    error.classList.remove('show');
    success.classList.remove('show');
}

// 로딩 숨기기
function hideLoading() {
    loading.classList.remove('show');
}

// 에러 메시지 표시
function showError(message) {
    error.textContent = message;
    error.classList.add('show');
}

// 성공 메시지 표시
function showSuccess(message) {
    success.textContent = message;
    success.classList.add('show');
}

// 이벤트 리스너 등록
loginBtn.addEventListener('click', login);
consoleBtn.addEventListener('click', launchConsole);
logoutBtn.addEventListener('click', logout);
