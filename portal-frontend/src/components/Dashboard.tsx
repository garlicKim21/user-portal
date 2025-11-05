import { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { 
  Menu,
  X,
  AlertCircle,
  Building2,
  LogOut
} from 'lucide-react';
import { backendAuthService } from '../services/backendAuthService';
import { ProjectSelector } from './ProjectSelector';
import { UserInfo } from './UserInfo';
import { AppUser, UserProject } from '../types/user';
import { verifyTerminalReady } from '../services/terminalService';

interface DashboardProps {
  user: AppUser;
  currentProject: UserProject | null;
  onProjectChange: (project: UserProject) => void;
  onLogout: () => void;
}

export function Dashboard({ user, currentProject, onProjectChange, onLogout }: DashboardProps) {
  const [activeMenu, setActiveMenu] = useState<string | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [isTitleClicked, setIsTitleClicked] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const menuItems = [
    {
      id: 'grafana',
      name: 'Grafana',
      logo: '/Grafana_logo.svg.png',
      url: 'https://grafana.miribit.cloud'
    },
    {
      id: 'terminal',
      name: 'Secure Web Terminal',
      logo: '/Kubernetes.png',
      url: 'https://portal.miribit.cloud/api/launch-console'
    },
    {
      id: 'jenkins',
      name: 'Jenkins',
      logo: '/jenkins.png',
      url: 'https://jenkins.miribit.cloud'
    },
    {
      id: 'argocd',
      name: 'ArgoCD',
      logo: '/Argo CD.png',
      url: 'https://argocd.miribit.cloud/auth/login?return_url=https://argocd.miribit.cloud/applications'
    }
  ];


  // 메시지 자동 숨김
  useEffect(() => {
    if (error || success) {
      const timer = setTimeout(() => {
        setError(null);
        setSuccess(null);
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [error, success]);

  const handleMenuClick = (menuId: string, url: string) => {
    setActiveMenu(menuId);
    // 메뉴 클릭 시에는 오른쪽 패널만 변경하고, 실제 실행은 버튼 클릭 시에만
  };

  const handleWebConsoleClick = async () => {
    if (!user?.access_token) {
      setError('인증 토큰이 없습니다. 다시 로그인해주세요.');
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      console.log('[Dashboard] Launching web console...');
      const result = await backendAuthService.launchWebConsole(user.access_token);
      const is_ready = await verifyTerminalReady(result.url);
      if (is_ready) {
        setSuccess('Web Console이 성공적으로 준비되었습니다! 새 탭에서 열립니다...');
        console.log('[Dashboard] Opening terminal in new tab:', result.url);
        
        const newWindow = window.open(result.url, '_blank');

        if (!newWindow) {
          setError('팝업이 차단되었습니다. 브라우저 설정에서 팝업을 허용해주세요.');
        }
      } else {
        setError('Web Terminal 생성이 실패하였거나 네트워크 문제로 접속할 수 없습니다. 관리자에게 연락 바랍니다.');
      }
    } catch (error) {
      console.error('[Dashboard] Failed to launch web console:', error);
      setError(`Web Console 실행에 실패했습니다: ${error instanceof Error ? error.message : '알 수 없는 오류'}`);
    } finally {
      setIsLoading(false);
    }
  };

  const handleTitleClick = () => {
    setIsTitleClicked(true);
    setActiveMenu(null);
    // 클릭 효과를 200ms 후에 리셋
    setTimeout(() => setIsTitleClicked(false), 200);
  };


  return (
    <div className="h-screen flex bg-background">
      {/* Sidebar */}
      <div className={`${sidebarOpen ? 'w-64' : 'w-16'} transition-all duration-300 bg-sidebar border-r border-sidebar-border flex flex-col`}>
        {/* Header */}
        <div className="p-4 border-b border-sidebar-border">
          <div className="flex items-center justify-between">
            {sidebarOpen && (
              <div 
                className={`flex items-center space-x-3 cursor-pointer hover:opacity-80 transition-all duration-200 ${
                  isTitleClicked ? 'scale-95 opacity-70' : 'hover:scale-105'
                }`}
                onClick={handleTitleClick}
              >
                <img 
                  src="/skhynix.png" 
                  alt="SK하이닉스" 
                  className="h-8 w-auto"
                />
                <h2 className="text-sm font-medium text-sidebar-foreground">빅데이터 분석 플랫폼</h2>
              </div>
            )}
            {!sidebarOpen && (
              <div 
                className={`flex justify-center w-full cursor-pointer hover:opacity-80 transition-all duration-200 ${
                  isTitleClicked ? 'scale-95 opacity-70' : 'hover:scale-105'
                }`}
                onClick={handleTitleClick}
              >
                <img 
                  src="/skhynix.png" 
                  alt="SK하이닉스" 
                  className="h-6 w-auto"
                />
              </div>
            )}
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
              return (
                <Button
                  key={item.id}
                  variant={activeMenu === item.id ? "default" : "ghost"}
                  className={`w-full justify-start ${sidebarOpen ? 'px-4' : 'px-2'} ${
                    activeMenu === item.id 
                      ? 'bg-sidebar-primary text-sidebar-primary-foreground' 
                      : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
                  }`}
                  onClick={() => handleMenuClick(item.id, item.url)}
                >
                  <img 
                    src={item.logo} 
                    alt={item.name} 
                    className={`h-5 w-5 ${sidebarOpen ? 'mr-3' : ''}`}
                  />
                  {sidebarOpen && <span>{item.name}</span>}
                </Button>
              );
            })}
          </nav>
        </div>

        {/* Current Project Info */}
        {sidebarOpen && currentProject && (
          <div className="p-4 border-t border-sidebar-border">
            <div className="text-xs text-sidebar-foreground/70 mb-1">현재 프로젝트</div>
            <div className="text-sm font-medium text-sidebar-foreground">{currentProject.name}</div>
            <div className="text-xs text-sidebar-foreground/70">{currentProject.roleLabel}</div>
          </div>
        )}
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        {/* Top Bar */}
        <header className="h-16 border-b border-border bg-card px-6 flex items-center justify-between">
          <div 
            className={`cursor-pointer hover:opacity-80 transition-all duration-200 ${
              isTitleClicked ? 'scale-95 opacity-70' : 'hover:scale-105'
            }`}
            onClick={handleTitleClick}
          >
            <h1 className="text-xl font-bold">Dashboard</h1>
          </div>
          <div className="flex items-center gap-4">
            {user.projects.length > 1 && (
              <ProjectSelector 
                projects={user.projects}
                currentProject={currentProject}
                onProjectChange={onProjectChange}
              />
            )}
            {user.projects.length === 1 && (
              <div className="flex items-center gap-2 px-3 py-2 bg-muted rounded-md">
                <Building2 className="h-4 w-4" />
                <span className="text-sm font-medium">{currentProject?.name}</span>
                <span className="text-xs px-2 py-1 bg-gray-200 rounded">
                  {currentProject?.roleLabel}
                </span>
              </div>
            )}
            <UserInfo 
              user={user}
            />
            <Button 
              variant="destructive" 
              size="sm"
              onClick={onLogout}
              className="flex items-center gap-2"
            >
              <LogOut className="h-4 w-4" />
              로그아웃
            </Button>
          </div>
        </header>

        {/* Content Area */}
        <main className="flex-1 p-6">
          {/* 상태 메시지 */}
          {(isLoading || error || success) && (
            <div className="mb-4">
              {isLoading && (
                <div className="bg-blue-50 border border-blue-200 text-blue-800 px-4 py-3 rounded-md flex items-center">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-3"></div>
                  처리 중...
                </div>
              )}
              {error && (
                <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-md flex items-center">
                  <AlertCircle className="h-4 w-4 mr-3" />
                  {error}
                </div>
              )}
              {success && (
                <div className="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-md flex items-center">
                  <div className="h-4 w-4 rounded-full bg-green-600 mr-3"></div>
                  {success}
                </div>
              )}
            </div>
          )}
          {activeMenu ? (
            <Card>
              <CardHeader>
                <div className="flex items-center space-x-4">
                  <img 
                    src={menuItems.find(item => item.id === activeMenu)?.logo} 
                    alt={menuItems.find(item => item.id === activeMenu)?.name}
                    className="h-12 w-auto"
                  />
                  <CardTitle>
                    {menuItems.find(item => item.id === activeMenu)?.name}
                  </CardTitle>
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">
                  {activeMenu === 'grafana' && 'Grafana 대시보드에 연결됩니다. 데이터 시각화 및 모니터링을 위한 도구입니다.'}
                  {activeMenu === 'terminal' && 'Secure Web Terminal에 연결됩니다. 안전한 웹 기반 터미널 환경을 제공합니다.'}
                  {activeMenu === 'jenkins' && 'Jenkins CI/CD 파이프라인에 연결됩니다. 빌드 및 배포 자동화를 관리합니다.'}
                  {activeMenu === 'argocd' && 'ArgoCD GitOps 플랫폼에 연결됩니다. Kubernetes 애플리케이션 배포를 관리합니다.'}
                </p>
                <div className="mt-4">
                  <Button
                    onClick={() => {
                      if (activeMenu === 'terminal') {
                        handleWebConsoleClick();
                      } else {
                        window.open(menuItems.find(item => item.id === activeMenu)?.url, '_blank');
                      }
                    }}
                    disabled={activeMenu === 'terminal' && (!user?.access_token || isLoading)}
                  >
                    {activeMenu === 'terminal' && isLoading ? (
                      <>
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-2"></div>
                        실행 중...
                      </>
                    ) : (
                      `${menuItems.find(item => item.id === activeMenu)?.name} ${activeMenu === 'terminal' ? '실행' : '열기'}`
                    )}
                  </Button>
                </div>
              </CardContent>
            </Card>
          ) : (
            <>
              {/* 환영 메시지와 사용자 정보 (전체 너비, 왼쪽 정렬) */}
              <Card className="w-full">
              <CardHeader>
                <CardTitle className="text-2xl font-bold text-left">
                  환영합니다! {user.family_name && user.given_name 
                    ? `${user.family_name}${user.given_name}` 
                    : user.name || user.preferred_username || '사용자'
                  }님!
                </CardTitle>
              </CardHeader>
              <CardContent>
                {/* 사용자 정보 (같은 라인에 배치) */}
                <div className="space-y-4">
                  <div className="flex items-center gap-4">
                    <span className="text-lg font-bold text-gray-700 min-w-[100px]">사용자 ID:</span>
                    <span className="text-lg font-semibold">{user.preferred_username || '정보 없음'}</span>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-lg font-bold text-gray-700 min-w-[100px]">사용자 이름:</span>
                    <span className="text-lg font-semibold">
                      {user.family_name && user.given_name 
                        ? `${user.family_name}${user.given_name}` 
                        : user.name || '정보 없음'
                      }
                    </span>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-lg font-bold text-gray-700 min-w-[100px]">이메일:</span>
                    <span className="text-lg font-semibold">{user.email || '정보 없음'}</span>
                  </div>
                </div>
              </CardContent>
              </Card>
            </>
          )}
        </main>
      </div>
    </div>
  );
}