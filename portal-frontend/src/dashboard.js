// 대시보드 페이지 전용 로직

let currentSession = null;

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

// 사용자 정보 표시
function showUserInfo(userId) {
    const currentPath = window.location.pathname;
    
    if (currentPath === '/dashboard') {
        userName.textContent = `Welcome, ${userId}!`;
    } else {
        window.location.href = '/dashboard';
    }
}

// 로그인 섹션 표시
function showLoginSection() {
    const currentPath = window.location.pathname;
    
    if (currentPath === '/dashboard') {
        window.location.href = '/';
    } else {
        loading.classList.remove('show');
        error.classList.remove('show');
        success.classList.remove('show');
    }
}

// 쿠키 가져오기
function getCookie(name) {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop().split(';').shift();
    return null;
}
let currentService = 'terminal';

// DOM 요소들
const consoleBtn = document.getElementById('consoleBtn');
const grafanaBtn = document.getElementById('grafanaBtn');
const argocdBtn = document.getElementById('argocdBtn');
const jenkinsBtn = document.getElementById('jenkinsBtn');
const logoutBtn = document.getElementById('logoutBtn');
const userName = document.getElementById('user-name');
const loading = document.getElementById('loading');
const error = document.getElementById('error');
const success = document.getElementById('success');

// 메뉴 요소들
const menuItems = document.querySelectorAll('.menu-item');
const contentTitle = document.getElementById('content-title');
const contentSubtitle = document.getElementById('content-subtitle');
const serviceContents = document.querySelectorAll('.service-content');

// 페이지 로드 시 사용자 상태 확인
document.addEventListener('DOMContentLoaded', function() {
    initializeDashboard();
});

// 대시보드 초기화 로직
async function initializeDashboard() {
    const isLoggedIn = await checkUserStatus();
    
    if (!isLoggedIn) {
        // 로그인되지 않은 경우 로그인 페이지로 리다이렉트
        window.location.href = '/';
        return;
    }
    
    // 사용자 정보 표시
    if (window.currentUser) {
        userName.textContent = `Welcome, ${window.currentUser.user_id}!`;
    }
    
    // 메뉴 이벤트 리스너 등록
    setupMenuListeners();
}

// 메뉴 이벤트 리스너 설정
function setupMenuListeners() {
    menuItems.forEach(item => {
        item.addEventListener('click', function(e) {
            e.preventDefault();
            const service = this.getAttribute('data-service');
            switchService(service);
        });
    });
}

// 서비스 전환
function switchService(service) {
    // 현재 활성 메뉴 제거
    menuItems.forEach(item => item.classList.remove('active'));
    
    // 새로운 활성 메뉴 설정
    const activeMenuItem = document.querySelector(`[data-service="${service}"]`);
    if (activeMenuItem) {
        activeMenuItem.classList.add('active');
    }
    
    // 모든 서비스 콘텐츠 숨기기
    serviceContents.forEach(content => {
        content.style.display = 'none';
    });
    
    // 선택된 서비스 콘텐츠 표시
    const targetContent = document.getElementById(`${service}-content`);
    if (targetContent) {
        targetContent.style.display = 'block';
    }
    
    // 콘텐츠 제목과 부제목 업데이트
    updateContentHeader(service);
    
    currentService = service;
}

// 콘텐츠 헤더 업데이트
function updateContentHeader(service) {
    const titles = {
        terminal: {
            title: 'Web Terminal',
            subtitle: 'Access your secure Kubernetes terminal'
        },
        grafana: {
            title: 'Grafana',
            subtitle: 'Monitor your cluster metrics and performance'
        },
        argocd: {
            title: 'ArgoCD',
            subtitle: 'GitOps-based continuous deployment'
        },
        jenkins: {
            title: 'Jenkins',
            subtitle: 'Continuous integration and deployment'
        }
    };
    
    const serviceInfo = titles[service];
    if (serviceInfo) {
        contentTitle.textContent = serviceInfo.title;
        contentSubtitle.textContent = serviceInfo.subtitle;
    }
}

// Grafana 실행
function launchGrafana() {
    try {
        console.log('Opening Grafana URL: https://grafana.miribit.cloud/');
        window.open('https://grafana.miribit.cloud/', '_blank');
    } catch (error) {
        showError(`Failed to open Grafana: ${error.message}`);
    }
}

// ArgoCD 실행
function launchArgoCD() {
    try {
        console.log('Opening ArgoCD URL: https://argocd.miribit.cloud/');
        window.open('https://argocd.miribit.cloud/', '_blank');
    } catch (error) {
        showError(`Failed to open ArgoCD: ${error.message}`);
    }
}

// Jenkins 실행
function launchJenkins() {
    try {
        console.log('Opening Jenkins URL: https://jenkins.miribit.cloud/');
        window.open('https://jenkins.miribit.cloud/', '_blank');
    } catch (error) {
        showError(`Failed to open Jenkins: ${error.message}`);
    }
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

// 웹 터미널 실행
async function launchConsole() {
    showLoading();

    try {
        const response = await fetch('/api/launch-console', {
            credentials: 'include'
        });
        
        if (response.ok) {
            const data = await response.json();
            console.log('Console response data:', data);
            hideLoading();
            showSuccess('Web terminal started successfully!');
            
            if (data.data && data.data.url) {
                console.log('Opening URL:', data.data.url);
                setTimeout(() => {
                    window.open(data.data.url, '_blank');
                }, 1000);
            } else {
                console.error('URL not found in response:', data);
                showError('Console URL not received from server');
            }
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
if (consoleBtn) {
    consoleBtn.addEventListener('click', launchConsole);
}
if (grafanaBtn) {
    grafanaBtn.addEventListener('click', launchGrafana);
}
if (argocdBtn) {
    argocdBtn.addEventListener('click', launchArgoCD);
}
if (jenkinsBtn) {
    jenkinsBtn.addEventListener('click', launchJenkins);
}
if (logoutBtn) {
    logoutBtn.addEventListener('click', logout);
}