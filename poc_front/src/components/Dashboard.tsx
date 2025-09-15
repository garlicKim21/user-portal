import { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { launchTerminal, deleteTerminal } from '../services/terminalService';
import { 
  BarChart3, 
  Terminal, 
  Wrench, 
  GitBranch, 
  LogOut,
  Menu,
  X,
  User,
  Mail,
  Calendar,
  Shield
} from 'lucide-react';

interface UserInfo {
  name?: string;
  email?: string;
  preferred_username?: string;
  given_name?: string;
  family_name?: string;
  roles?: string[];
  groups?: string[];
  sub?: string;
  iat?: number;
  exp?: number;
  [key: string]: any;
}

interface DashboardProps {
  onLogout: () => void;
  user?: UserInfo; // OIDC 사용자 정보
}

export function Dashboard({ onLogout, user }: DashboardProps) {
  const [activeMenu, setActiveMenu] = useState<string | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);

  // 디버깅을 위한 사용자 정보 로그
  useEffect(() => {
    console.log('Dashboard user data:', user);
  }, [user]);

  const menuItems = [
    {
      id: 'grafana',
      name: 'Grafana',
      icon: BarChart3,
      iconImage: '/grafana.svg',
      url: 'https://grafana.miribit.cloud/auth/login?return_url=https://grafana.miribit.cloud/applications'
    },
    {
      id: 'terminal',
      name: 'Secure Web Terminal',
      icon: Terminal,
      iconImage: null,
      url: '#terminal'
    },
    {
      id: 'jenkins',
      name: 'Jenkins',
      icon: Wrench,
      iconImage: '/jenkins.png',
      url: 'https://jenkins.miribit.cloud'
    },
    {
      id: 'argocd',
      name: 'ArgoCD',
      icon: GitBranch,
      iconImage: '/argocd.png',
      url: 'https://argocd.miribit.cloud/auth/login?return_url=https://argocd.miribit.cloud/applications'
    }
  ];

  const handleMenuClick = (menuId: string) => {
    setActiveMenu(menuId);
  };

  // 터미널 팝업 창 열기 함수
  const openTerminalPopup = async () => {
    try {
      const result = await launchTerminal();

      // 팝업 창 열기
      const popup = window.open(
        '',
        'web-terminal',
        'width=1200,height=800,scrollbars=yes,resizable=yes,toolbar=no,menubar=no,location=no,status=no'
      );

      if (!popup) {
        alert('팝업이 차단되었습니다. 팝업 차단을 해제해주세요.');
        return;
      }

      // 팝업 창에 HTML 내용 작성
      popup.document.write(`
        <!DOCTYPE html>
        <html lang="ko">
        <head>
          <meta charset="UTF-8">
          <meta name="viewport" content="width=device-width, initial-scale=1.0">
          <title>Secure Web Terminal</title>
          <style>
            * { margin: 0; padding: 0; box-sizing: border-box; }
            body { 
              font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
              background: #0a0a0a;
              color: #ffffff;
              overflow: hidden;
            }
            .header {
              background: #1a1a1a;
              padding: 12px 16px;
              border-bottom: 1px solid #333;
              display: flex;
              align-items: center;
              justify-content: space-between;
            }
            .header h1 {
              font-size: 16px;
              font-weight: 600;
              display: flex;
              align-items: center;
              gap: 8px;
            }
            .status {
              font-size: 12px;
              color: #888;
              background: #333;
              padding: 4px 8px;
              border-radius: 4px;
            }
            .terminal-container {
              height: calc(100vh - 60px);
              position: relative;
            }
            iframe {
              width: 100%;
              height: 100%;
              border: none;
              background: #000;
            }
            .loading {
              position: absolute;
              top: 50%;
              left: 50%;
              transform: translate(-50%, -50%);
              text-align: center;
              color: #888;
            }
            .spinner {
              width: 32px;
              height: 32px;
              border: 3px solid #333;
              border-top: 3px solid #007acc;
              border-radius: 50%;
              animation: spin 1s linear infinite;
              margin: 0 auto 16px;
            }
            @keyframes spin {
              0% { transform: rotate(0deg); }
              100% { transform: rotate(360deg); }
            }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>🔒 Secure Web Terminal</h1>
            <div class="status" id="status">연결 중...</div>
          </div>
          <div class="terminal-container">
            <div class="loading" id="loading">
              <div class="spinner"></div>
              <div>터미널을 시작하는 중...</div>
            </div>
            <iframe id="terminal" src="${result.url}" style="display: none;"></iframe>
          </div>
          <script>
            const iframe = document.getElementById('terminal');
            const loading = document.getElementById('loading');
            const status = document.getElementById('status');
            
            iframe.onload = function() {
              loading.style.display = 'none';
              iframe.style.display = 'block';
              status.textContent = '연결됨';
            };
            
            iframe.onerror = function() {
              loading.innerHTML = '<div style="color: #ff6b6b;">터미널 연결에 실패했습니다.</div>';
              status.textContent = '연결 실패';
            };
            
            // 팝업 창이 닫힐 때 부모 창에 알림
            window.addEventListener('beforeunload', function() {
              if (window.opener) {
                window.opener.postMessage({ type: 'terminal-closed', resourceId: '${result.resourceId}' }, '*');
              }
            });
          </script>
        </body>
        </html>
      `);

      popup.document.close();

      // 팝업 창이 닫혔는지 확인하는 인터벌
      const checkClosed = setInterval(() => {
        if (popup.closed) {
          clearInterval(checkClosed);
        }
      }, 1000);

    } catch (error) {
      console.error('Terminal launch error:', error);
      alert(`터미널 시작에 실패했습니다: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleOpenService = (menuId: string, url: string) => {
    if (menuId === 'terminal') {
      // 터미널의 경우 팝업 창으로 열기
      openTerminalPopup();
    } else {
      // 다른 메뉴의 경우 새 탭에서 열기
      window.open(url, '_blank');
    }
  };

  // 팝업 창에서 메시지 받기
  useEffect(() => {
    const handleMessage = async (event: MessageEvent) => {
      if (event.data.type === 'terminal-closed' && event.data.resourceId) {
        try {
          await deleteTerminal(event.data.resourceId);
        } catch (error) {
          console.error('Failed to cleanup terminal:', error);
        }
      }
    };

    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, []);

  // 사용자 정보 포맷팅 함수
  const formatUserInfo = (user: UserInfo) => {
    const displayName = user.name || user.preferred_username || user.given_name || '사용자';
    const email = user.email || '';
    const roles = user.roles || user.groups || [];
    const loginTime = user.iat ? new Date(user.iat * 1000).toLocaleString('ko-KR') : '';
    
    return {
      displayName,
      email,
      roles,
      loginTime
    };
  };

  return (
    <div className="h-screen flex bg-background">
      {/* Sidebar */}
      <div className={`${sidebarOpen ? 'w-64' : 'w-16'} transition-all duration-300 bg-sidebar border-r border-sidebar-border flex flex-col`}>
        {/* Header */}
        <div className="p-4 border-b border-sidebar-border">
          <div className="flex items-center justify-between">
            <div className="flex flex-col items-center space-y-2">
              <img 
                src="/skhynix.png" 
                alt="SK하이닉스 로고" 
                className="h-8 w-auto cursor-pointer hover:opacity-80 transition-opacity"
                onClick={() => setActiveMenu(null)}
              />
              {sidebarOpen && (
                <div className="text-center">
                  <h2 className="text-sidebar-foreground text-sm font-medium">빅데이터 분석 플랫폼</h2>
                  <p className="text-sidebar-foreground text-xs text-muted-foreground">사용자 포털</p>
                </div>
              )}
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="text-sidebar-foreground hover:bg-sidebar-accent"
            >
              {sidebarOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
            </Button>
          </div>
        </div>

        {/* Menu Items */}
        <div className="flex-1 p-2">
          <nav className="space-y-2">
            {menuItems.map((item) => {
              const Icon = item.icon;
              return (
                <Button
                  key={item.id}
                  variant={activeMenu === item.id ? "default" : "ghost"}
                  className={`w-full justify-start ${sidebarOpen ? 'px-4' : 'px-2'} ${
                    activeMenu === item.id 
                      ? 'bg-sidebar-primary text-sidebar-primary-foreground' 
                      : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
                  }`}
                  onClick={() => handleMenuClick(item.id)}
                >
                  {item.iconImage ? (
                    <img 
                      src={item.iconImage} 
                      alt={`${item.name} 아이콘`} 
                      className={`h-5 w-5 ${sidebarOpen ? 'mr-3' : ''}`}
                    />
                  ) : (
                    <Icon className={`h-5 w-5 ${sidebarOpen ? 'mr-3' : ''}`} />
                  )}
                  {sidebarOpen && <span>{item.name}</span>}
                </Button>
              );
            })}
          </nav>
        </div>

        {/* Logout Button */}
        <div className="p-2 border-t border-sidebar-border">
          <Button
            variant="ghost"
            className={`w-full justify-start ${sidebarOpen ? 'px-4' : 'px-2'} text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground`}
            onClick={onLogout}
          >
            <LogOut className={`h-5 w-5 ${sidebarOpen ? 'mr-3' : ''}`} />
            {sidebarOpen && <span>로그아웃</span>}
          </Button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        {/* Top Bar */}
        <header className="h-16 border-b border-border bg-card px-6 flex items-center justify-between">
          <div>
            <h1>대시보드</h1>
            <p className="text-muted-foreground">개발자 및 빅데이터 분석가를 위한 포털</p>
          </div>
        </header>

        {/* Content Area */}
        <main className="flex-1 p-6">
          {activeMenu ? (
            <Card>
              <CardHeader>
                <div className="flex items-center">
                  {menuItems.find(item => item.id === activeMenu)?.iconImage && (
                    <img 
                      src={menuItems.find(item => item.id === activeMenu)?.iconImage || ''} 
                      alt={`${menuItems.find(item => item.id === activeMenu)?.name || ''} 로고`} 
                      className="h-16 w-auto mr-6"
                    />
                  )}
                  <CardTitle className="text-3xl font-bold mr-6">
                    {menuItems.find(item => item.id === activeMenu)?.name || ''}
                  </CardTitle>
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">
                  {activeMenu === 'grafana' && 'Grafana 대시보드에 연결됩니다. 데이터 시각화 및 모니터링을 위한 도구입니다.'}
                  {activeMenu === 'terminal' && 'Secure Web Terminal을 팝업으로 열어 안전한 웹 기반 터미널 환경을 제공합니다.'}
                  {activeMenu === 'jenkins' && 'Jenkins CI/CD 파이프라인에 연결됩니다. 빌드 및 배포 자동화를 관리합니다.'}
                  {activeMenu === 'argocd' && 'ArgoCD GitOps 플랫폼에 연결됩니다. Kubernetes 애플리케이션 배포를 관리합니다.'}
                </p>
                <div className="mt-6">
                  <Button
                    size="lg"
                    onClick={() => {
                      const selectedItem = menuItems.find(item => item.id === activeMenu);
                      if (selectedItem) {
                        handleOpenService(selectedItem.id, selectedItem.url);
                      }
                    }}
                  >
                    {menuItems.find(item => item.id === activeMenu)?.name} 열기
                  </Button>
                </div>
              </CardContent>
            </Card>
          ) : (
            <div className="flex flex-col lg:flex-row gap-6 h-full">
              {/* 사용자 정보 카드 */}
              <Card className="lg:w-96 h-fit">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <User className="h-5 w-5" />
                    사용자 정보
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {user ? (() => {
                    const userInfo = formatUserInfo(user);
                    return (
                      <div className="space-y-4">
                        <div className="flex items-center gap-3">
                          <div className="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center">
                            <User className="h-6 w-6 text-primary" />
                          </div>
                          <div>
                            <h3 className="font-semibold text-lg">{userInfo.displayName}</h3>
                            {userInfo.email && (
                              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                <Mail className="h-4 w-4" />
                                {userInfo.email}
                              </div>
                            )}
                          </div>
                        </div>
                        
                        {userInfo.roles.length > 0 && (
                          <div>
                            <h4 className="text-sm font-medium mb-2 flex items-center gap-2">
                              <Shield className="h-4 w-4" />
                              권한 및 역할
                            </h4>
                            <div className="flex flex-wrap gap-2">
                              {userInfo.roles.map((role, index) => (
                                <span
                                  key={index}
                                  className="px-2 py-1 bg-secondary text-secondary-foreground text-xs rounded-md"
                                >
                                  {role}
                                </span>
                              ))}
                            </div>
                          </div>
                        )}
                        
                        {userInfo.loginTime && (
                          <div className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Calendar className="h-4 w-4" />
                            로그인 시간: {userInfo.loginTime}
                          </div>
                        )}
                      </div>
                    );
                  })() : (
                    <div className="space-y-4">
                      <div className="flex items-center gap-3">
                        <div className="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center">
                          <User className="h-6 w-6 text-primary" />
                        </div>
                        <div>
                          <h3 className="font-semibold text-lg">사용자</h3>
                          <div className="text-sm text-muted-foreground">
                            사용자 정보를 불러오는 중...
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>

              {/* 서비스 안내 카드 */}
              <Card className="flex-1">
                <CardHeader>
                  <CardTitle>환영합니다!</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-muted-foreground mb-6">
                    왼쪽 메뉴에서 원하시는 도구를 선택하여 시작하세요.
                  </p>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="flex items-center p-3 rounded-lg border bg-card">
                      <img 
                        src="/grafana.svg" 
                        alt="Grafana 로고" 
                        className="h-8 w-8 mr-3"
                      />
                      <div>
                        <h3 className="font-medium">Grafana</h3>
                        <p className="text-sm text-muted-foreground">데이터 시각화</p>
                      </div>
                    </div>
                    <div className="flex items-center p-3 rounded-lg border bg-card">
                      <Terminal className="h-8 w-8 mr-3 text-green-500" />
                      <div>
                        <h3 className="font-medium">Secure Web Terminal</h3>
                        <p className="text-sm text-muted-foreground">안전한 터미널</p>
                      </div>
                    </div>
                    <div className="flex items-center p-3 rounded-lg border bg-card">
                      <img 
                        src="/jenkins.png" 
                        alt="Jenkins 로고" 
                        className="h-8 w-8 mr-3"
                      />
                      <div>
                        <h3 className="font-medium">Jenkins</h3>
                        <p className="text-sm text-muted-foreground">CI/CD 파이프라인</p>
                      </div>
                    </div>
                    <div className="flex items-center p-3 rounded-lg border bg-card">
                      <img 
                        src="/argocd.png" 
                        alt="ArgoCD 로고" 
                        className="h-8 w-8 mr-3"
                      />
                      <div>
                        <h3 className="font-medium">ArgoCD</h3>
                        <p className="text-sm text-muted-foreground">GitOps 배포</p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}
        </main>
      </div>

    </div>
  );
}