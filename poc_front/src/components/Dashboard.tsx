import { useState } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { WebTerminal } from './WebTerminal';
import { 
  BarChart3, 
  Terminal, 
  Wrench, 
  GitBranch, 
  LogOut,
  Menu,
  X
} from 'lucide-react';

interface DashboardProps {
  onLogout: () => void;
  user?: any; // OIDC 사용자 정보
}

export function Dashboard({ onLogout, user }: DashboardProps) {
  const [activeMenu, setActiveMenu] = useState<string | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [isTerminalOpen, setIsTerminalOpen] = useState(false);

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

  const handleOpenService = (menuId: string, url: string) => {
    if (menuId === 'terminal') {
      // 터미널의 경우 팝업으로 열기
      setIsTerminalOpen(true);
    } else {
      // 다른 메뉴의 경우 새 탭에서 열기
      window.open(url, '_blank');
    }
  };

  return (
    <div className="h-screen flex bg-background">
      {/* Sidebar */}
      <div className={`${sidebarOpen ? 'w-64' : 'w-16'} transition-all duration-300 bg-sidebar border-r border-sidebar-border flex flex-col`}>
        {/* Header */}
        <div className="p-4 border-b border-sidebar-border">
          <div className="flex items-center justify-between">
            {sidebarOpen && (
              <div className="flex flex-col items-center space-y-2">
                <img 
                  src="/skhynix.png" 
                  alt="SK하이닉스 로고" 
                  className="h-8 w-auto"
                />
                <h2 className="text-sidebar-foreground text-center text-sm">빅데이터 분석 플랫폼 사용자 포털</h2>
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
            <div className="flex items-center justify-center h-full">
              <Card className="max-w-md">
                <CardHeader>
                  <CardTitle>환영합니다!</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-muted-foreground">
                    왼쪽 메뉴에서 원하시는 도구를 선택하여 시작하세요.
                  </p>
                  <div className="mt-4 space-y-2">
                    <div className="flex items-center text-sm text-muted-foreground">
                      <BarChart3 className="h-4 w-4 mr-2" />
                      Grafana - 데이터 시각화
                    </div>
                    <div className="flex items-center text-sm text-muted-foreground">
                      <Terminal className="h-4 w-4 mr-2" />
                      Secure Web Terminal - 안전한 터미널
                    </div>
                    <div className="flex items-center text-sm text-muted-foreground">
                      <Wrench className="h-4 w-4 mr-2" />
                      Jenkins - CI/CD 파이프라인
                    </div>
                    <div className="flex items-center text-sm text-muted-foreground">
                      <GitBranch className="h-4 w-4 mr-2" />
                      ArgoCD - GitOps 배포
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          )}
        </main>
      </div>

      {/* Web Terminal Popup */}
      <WebTerminal 
        isOpen={isTerminalOpen}
        onClose={() => setIsTerminalOpen(false)}
        user={user}
      />
    </div>
  );
}