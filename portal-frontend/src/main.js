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
    console.log('페이지 로드됨');
    checkCookies(); // 쿠키 디버깅
    // handleRouting();
    initializeApp();
});

// 앱 초기화 로직
async function initializeApp() {
    const isLoggedIn = await checkUserStatus(); // 먼저 로그인 상태 확인
    const path = window.location.pathname;

    console.log('현재 경로:', path, '로그인 상태:', isLoggedIn);

    if (isLoggedIn) {
        // 로그인 상태라면, 어느 경로에 있든 대시보드로 보냄
        if (path !== '/dashboard') {
            // 주소창의 URL도 변경해줌 (깜빡임 방지)
            window.history.pushState({}, '', '/dashboard');
        }
        showDashboardPage(true); // 사용자 정보를 이미 아므로 바로 표시
    } else {
        // 로그아웃 상태라면, 어느 경로에 있든 로그인 페이지로 보냄
        if (path !== '/') {
           window.history.pushState({}, '', '/');
        }
        showLoginPage();
    }
}

// checkUserStatus 함수는 로그인 여부(true/false)를 반환하도록 수정
async function checkUserStatus() {
    try {
        const response = await fetch('/api/user', { credentials: 'include' });
        if (response.ok) {
            const userData = await response.json();
            console.log('사용자 데이터:', userData);
            // 사용자 정보를 어딘가에 저장해두면 좋음
            window.currentUser = userData; 
            return true; // 로그인 성공
        }
        return false; // 로그인 실패
    } catch (error) {
        console.error('Error checking user status:', error);
        return false;
    }
}

// 페이지 표시 함수들은 파라미터를 받아서 UI만 제어하도록 단순화
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

// 주석처리
// // 라우팅 처리
// function handleRouting() {
//     const path = window.location.pathname;
//     console.log('현재 경로:', path);
    
//     if (path === '/dashboard') {
//         showDashboardPage();
//     } else {
//         showLoginPage();
//     }
// }

// // 대시보드 페이지 표시
// function showDashboardPage() {
//     loginPage.style.display = 'none';
//     dashboardPage.style.display = 'block';
//     checkUserStatus();
// }

// // 로그인 페이지 표시
// function showLoginPage() {
//     loginPage.style.display = 'block';
//     dashboardPage.style.display = 'none';
//     checkUserStatus();
// }

// // 사용자 상태 확인
// async function checkUserStatus() {
//     try {
//         // 쿠키 디버깅
//         console.log('현재 쿠키들:', document.cookie);
        
//         // JWT 기반 세션 확인 (쿠키에서 자동으로 전송됨)
//         const response = await fetch('/api/user', {
//             credentials: 'include' // 쿠키 포함
//         });
        
//         console.log('API 응답 상태:', response.status);
        
//         if (response.ok) {
//             const userData = await response.json();
//             console.log('사용자 데이터:', userData);
//             showUserInfo(userData.user_id);
//         } else {
//             console.log('API 응답 실패:', response.status, response.statusText);
//             showLoginSection();
//         }
//     } catch (error) {
//         console.error('Error checking user status:', error);
//         showLoginSection();
//     }
// }

// 로그인
function login() {
    window.location.href = '/api/login';
}

// 로그아웃
// 로그아웃
async function logout() {
    try {
        // 백엔드의 로그아웃 엔드포인트를 호출하여 서버 측에서 쿠키를 삭제합니다.
        const response = await fetch('/api/logout', {
            method: 'POST', // 로그아웃은 상태를 변경하므로 POST가 적합합니다.
        });

        if (response.ok) {
            console.log('Successfully logged out from server');
        } else {
            console.error('Server logout failed');
        }
    } catch (error) {
        console.error('Logout request error:', error);
    } finally {
        // 서버 요청 성공 여부와 관계없이 클라이언트는 로그인 페이지로 이동합니다.
        window.location.href = '/';
    }
}

// 사용자 정보 표시
function showUserInfo(userId) {
    const currentPath = window.location.pathname;
    
    if (currentPath === '/dashboard') {
        // 대시보드 페이지에서는 사용자 정보 표시
        userName.textContent = `Welcome, ${userId}!`;
        userInfo.classList.add('show');
    } else {
        // 로그인 페이지에서는 대시보드로 리디렉션
        window.location.href = '/dashboard';
    }
}

// 로그인 섹션 표시
function showLoginSection() {
    const currentPath = window.location.pathname;
    
    if (currentPath === '/dashboard') {
        // 대시보드 페이지에서 로그인되지 않은 경우 로그인 페이지로 리디렉션
        window.location.href = '/';
    } else {
        // 로그인 페이지에서는 로그인 버튼 표시
        loginBtn.style.display = 'inline-block';
        userInfo.classList.remove('show');
        loading.classList.remove('show');
        error.classList.remove('show');
        success.classList.remove('show');
    }
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

// 쿠키 확인 함수
function checkCookies() {
    console.log('=== 쿠키 디버깅 ===');
    console.log('전체 쿠키:', document.cookie);
    
    const cookies = document.cookie.split(';');
    cookies.forEach(cookie => {
        const [name, value] = cookie.trim().split('=');
        console.log(`쿠키: ${name} = ${value ? value.substring(0, 20) + '...' : 'empty'}`);
    });
    
    // 새로운 쿠키명으로 확인
    const sessionToken = getCookie('portal-session');
    console.log('session_token 쿠키:', sessionToken ? '존재함' : '없음');
    
    // 쿠키가 없는 경우 추가 디버깅
    if (!sessionToken) {
        console.log('쿠키가 없는 이유 분석:');
        console.log('- 현재 도메인:', window.location.hostname);
        console.log('- 현재 프로토콜:', window.location.protocol);
        console.log('- User Agent:', navigator.userAgent);
    }
}

// 쿠키 값 가져오기
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
    handleRouting();
});
