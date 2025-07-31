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

// URL에서 세션 파라미터 확인 (JWT 기반으로 변경되어 더 이상 사용하지 않음)
// function getSessionFromURL() {
//     return null;
// }

// 사용자 상태 확인
async function checkUserStatus() {
    try {
        // JWT 기반 세션 확인 (쿠키에서 자동으로 전송됨)
        const response = await fetch('/api/user', {
            credentials: 'include' // 쿠키 포함
        });
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
}

// 로그인
function login() {
    window.location.href = '/api/login';
}

// 로그아웃
async function logout() {
    try {
        // 쿠키 삭제를 위한 서버 요청 (추후 구현 필요)
        document.cookie = 'session_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
        showLoginSection();
        // URL에서 세션 파라미터 제거
        window.history.replaceState({}, document.title, window.location.pathname);
    } catch (error) {
        console.error('Logout error:', error);
        showLoginSection();
    }
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
    showLoading();

    try {
        const response = await fetch('/api/launch-console', {
            credentials: 'include' // 쿠키 포함
        });
        
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
