import { useState, useEffect } from 'react';
import { useAuth } from 'react-oidc-context';
import { LoginPage } from './LoginPage';
import { Dashboard } from './Dashboard';
import { Callback } from './Callback';
import { useLocation } from 'react-router-dom';
import { AppUser, UserProject, AuthState, mockProjects } from '../types/user';
import env, { getOidcStorageKey, getKeycloakEndpoints } from '../config/env';

// Keycloak groups에서 프로젝트 정보 파싱 함수
// 새로운 그룹 형식 지원:
// - 그룹은 '/'로 계층화되고 깊이는 가변
// - 마지막 바로 앞 토큰이 네임스페이스
// - 마지막 토큰은 언더바('_')로 구분되며, 언더바 뒤가 실제 권한 키(adm|dev|view)
function parseUserProjectsFromGroups(groups: string[] | undefined): UserProject[] {
  console.log('parseUserProjectsFromGroups called with groups:', groups);
  
  if (!groups || groups.length === 0) {
    console.log('No groups found, returning empty array');
    return [];
  }

  const projects: UserProject[] = [];
  const projectMap = new Map<string, UserProject>(); // 중복 제거를 위한 맵
  
  groups.forEach(group => {
    console.log('Processing group:', group);
    
    // 그룹이 '/'로 시작하지 않으면 무시
    if (!group.startsWith('/')) {
      console.log('Group does not start with /:', group);
      return;
    }

    // '/'로 split (깊이 가변)
    const parts = group.substring(1).split('/');
    if (parts.length < 2) {
      console.log('Group does not have enough depth:', group);
      return;
    }

    // 마지막 바로 앞 토큰이 네임스페이스
    const namespace = parts[parts.length - 2];
    // 마지막 토큰에서 언더바 뒤가 권한 키
    const lastToken = parts[parts.length - 1];
    
    const underscoreIdx = lastToken.lastIndexOf('_');
    if (underscoreIdx === -1 || underscoreIdx === lastToken.length - 1) {
      console.log('Last token does not have underscore format:', lastToken);
      return;
    }

    const roleKey = lastToken.substring(underscoreIdx + 1);
    
    // 유효한 권한 키인지 확인 (adm, dev, view)
    if (roleKey !== 'adm' && roleKey !== 'dev' && roleKey !== 'view') {
      console.log('Invalid role key:', roleKey);
      return;
    }

    console.log('Parsed namespace:', namespace, 'role:', roleKey);
    
    // 프로젝트 ID를 기반으로 프로젝트명 매핑
    const projectName = getProjectName(namespace);
    // roleKey가 'view'인 경우 'viewer'로 변환 (프론트엔드 타입 호환)
    const role = roleKey === 'view' ? 'viewer' : (roleKey as 'dev' | 'adm');
    const roleLabel = getRoleLabel(roleKey);
    
    // 동일한 네임스페이스가 이미 있으면, 더 높은 권한으로 업데이트
    // 우선순위: adm > dev > view
    const existingProject = projectMap.get(namespace);
    if (existingProject) {
      const priority = { 'adm': 1, 'dev': 2, 'view': 3, 'viewer': 3 };
      const existingPriority = priority[existingProject.role] || 99;
      const newPriority = priority[role] || 99;
      
      if (newPriority < existingPriority) {
        // 더 높은 권한으로 업데이트
        projectMap.set(namespace, {
          id: namespace,
          name: projectName,
          role: role,
          roleLabel: roleLabel
        });
      }
    } else {
      projectMap.set(namespace, {
        id: namespace,
        name: projectName,
        role: role,
        roleLabel: roleLabel
      });
    }
  });

  // Map의 값들을 배열로 변환
  const finalProjects = Array.from(projectMap.values());
  console.log('Final projects array:', finalProjects);
  return finalProjects;
}

// 프로젝트 ID를 프로젝트명으로 매핑 (동적 생성)
function getProjectName(projectId: string): string {
  // 프로젝트 ID를 대문자로 변환하고 '프로젝트' 추가
  return `${projectId.toUpperCase()} 프로젝트`;
}

// 역할 코드를 역할명으로 매핑
function getRoleLabel(roleKey: string): string {
  const roleLabels: Record<string, string> = {
    'dev': '개발자',
    'adm': '관리자',
    'view': '조회자',
    'viewer': '조회자' // 호환성을 위해 유지
  };
  return roleLabels[roleKey] || roleKey;
}

export function AuthWrapper() {
  const { signinRedirect, isAuthenticated, user, isLoading, error } = useAuth();
  const location = useLocation();
  
  // === 디버깅 로그 ===
  console.log('🟡 [AuthWrapper] 렌더링:', {
    isAuthenticated,
    isLoading,
    hasUser: !!user,
    pathname: location.pathname,
    hasError: !!error
  });
  
  if (error) {
    console.error('🔴 [AuthWrapper] OIDC 에러 발생:', error);
    console.error('🔴 [AuthWrapper] 에러 상세:', JSON.stringify(error, null, 2));
  }
  
  // 상태 관리
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    currentProject: null
  });

  // 사용자 로그인 시 실제 OIDC 토큰 정보와 프로젝트 데이터 결합
  useEffect(() => {
    if (isAuthenticated && user && !authState.user) {
      console.log('OIDC User Info:', user); // 디버깅용
      console.log('OIDC Profile:', user.profile); // 디버깅용
      console.log('OIDC Groups from profile:', user.profile?.groups); // 디버깅용
      console.log('OIDC Groups from user:', (user as any)?.groups); // 디버깅용
      
      // 실제 OIDC 사용자 정보를 AppUser 타입으로 변환
      const profile = user.profile || {};
      const appUserData: AppUser = {
        // OIDC 기본 정보 - profile 객체에서 가져오기
        sub: (profile as any)?.sub || (user as any)?.sub || '',
        preferred_username: (profile as any)?.preferred_username || (user as any)?.preferred_username || '',
        name: (profile as any)?.name || (user as any)?.name || 'Unknown User',
        email: (profile as any)?.email || (user as any)?.email || '',
        given_name: (profile as any)?.given_name || (user as any)?.given_name || '',
        family_name: (profile as any)?.family_name || (user as any)?.family_name || '',
        email_verified: (profile as any)?.email_verified || (user as any)?.email_verified || false,
        auth_time: (profile as any)?.auth_time || (user as any)?.auth_time || 0,
        access_token: (user as any)?.access_token || '',
        id_token: (user as any)?.id_token || '',
        
        // 앱에서 관리하는 프로젝트 정보
        // 실제 환경에서는 Keycloak groups 필드에서 파싱하거나 LDAP API 호출
        projects: parseUserProjectsFromGroups((profile as any)?.groups || (user as any)?.groups) || mockProjects
      };
      
      console.log('App User Data:', appUserData); // 디버깅용
      
      setAuthState({
        user: appUserData,
        currentProject: appUserData.projects[0] || null // 첫 번째 프로젝트를 기본값으로 설정
      });
    }
  }, [isAuthenticated, user, authState.user]);

  // 프로젝트 변경 핸들러
  const handleProjectChange = (project: UserProject) => {
    setAuthState(prev => ({
      ...prev,
      currentProject: project
    }));
  };

  // 로그아웃 핸들러 (웹 콘솔 리소스 삭제 + 키클락 세션 로그아웃)
  const handleLogout = async () => {
    try {
      console.log('로그아웃 시작...');
      
      // 1. 백엔드 리소스 정리 API 호출 (웹 콘솔 리소스 삭제)
      if (authState.user?.access_token) {
        try {
          const { backendAuthService } = await import('../services/backendAuthService');
          await backendAuthService.logout(authState.user.access_token);
          console.log('웹 콘솔 리소스 정리 완료');
        } catch (error) {
          console.error('웹 콘솔 리소스 정리 실패:', error);
          // 리소스 정리 실패해도 로그아웃은 계속 진행
        }
      }
      
      // 2. 로컬 스토리지에서 OIDC 사용자 정보 정리
      const oidcUserKey = getOidcStorageKey();
      localStorage.removeItem(oidcUserKey);
      sessionStorage.clear();
      
      // 3. 상태 초기화
      setAuthState({
        user: null,
        currentProject: null
      });
      
      // 4. 키클락 세션 로그아웃 (확인 페이지 없이 바로 로그아웃)
      const idToken = user?.id_token;
      const keycloakEndpoints = getKeycloakEndpoints();
      let logoutUrl;
      
      if (idToken) {
        // id_token_hint 사용으로 확인 페이지 생략
        logoutUrl = `${keycloakEndpoints.endSession}?client_id=${env.KEYCLOAK_CLIENT_ID}&post_logout_redirect_uri=${encodeURIComponent(env.PORTAL_URL)}&id_token_hint=${idToken}`;
      } else {
        logoutUrl = `${keycloakEndpoints.endSession}?client_id=${env.KEYCLOAK_CLIENT_ID}&post_logout_redirect_uri=${encodeURIComponent(env.PORTAL_URL)}`;
      }
      
      console.log('키클락 로그아웃 URL로 리다이렉트:', logoutUrl);
      window.location.href = logoutUrl;
      
    } catch (error) {
      console.error('로그아웃 중 오류:', error);
      // 오류가 발생해도 강제로 홈으로 리다이렉트
      window.location.href = '/';
    }
  };

  // 로딩 중일 때
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">인증 확인 중...</p>
        </div>
      </div>
    );
  }

  // 콜백 페이지인 경우
  if (location.pathname === '/callback') {
    return <Callback />;
  }

  // 인증되지 않은 경우 로그인 페이지 표시
  if (!isAuthenticated) {
    const handleLogin = async () => {
      try {
        console.log('🟡 [AuthWrapper] signinRedirect() 호출 시작');
        console.log('🟡 [AuthWrapper] 현재 URL:', window.location.href);
        await signinRedirect();
        console.log('🟡 [AuthWrapper] signinRedirect() 호출 완료');
      } catch (err) {
        console.error('🔴 [AuthWrapper] signinRedirect() 에러:', err);
        console.error('🔴 [AuthWrapper] 에러 상세:', JSON.stringify(err, null, 2));
        throw err;
      }
    };
    return <LoginPage onLogin={handleLogin} />;
  }

  // 인증된 경우 대시보드 표시 (사용자 정보가 로드되지 않은 경우 로딩 표시)
  if (!authState.user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">사용자 정보를 불러오는 중...</p>
        </div>
      </div>
    );
  }

  return (
    <Dashboard 
      user={authState.user}
      currentProject={authState.currentProject}
      onProjectChange={handleProjectChange}
      onLogout={handleLogout}
    />
  );
}