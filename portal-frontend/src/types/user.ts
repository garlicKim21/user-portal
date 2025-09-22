// 사용자 및 프로젝트 관련 타입 정의

export interface UserProject {
  id: string;        // 프로젝트 고유 ID (예: 'fdc', 'eds')
  name: string;      // 프로젝트 표시명 (예: 'FDC 프로젝트')
  role: 'dev' | 'adm' | 'viewer'; // LDAP 권한 코드
  roleLabel: string; // 권한 표시명 (예: '개발자', '관리자')
}

// 실제 OIDC 사용자 정보를 확장한 타입
export interface AppUser {
  // OIDC 기본 정보
  sub: string;
  preferred_username: string;
  name: string;
  email?: string;
  given_name?: string;
  family_name?: string;
  email_verified?: boolean;
  auth_time?: number;
  access_token?: string;
  id_token?: string;
  
  // 앱에서 추가로 관리하는 정보
  projects: UserProject[]; // 소속 프로젝트 목록
}

export interface AuthState {
  user: AppUser | null;
  currentProject: UserProject | null;
}

// Mock 프로젝트 데이터 (실제 환경에서는 LDAP에서 가져옴)
export const mockProjects: UserProject[] = [
  { id: 'fdc', name: 'FDC 프로젝트', role: 'dev', roleLabel: '개발자' },
  { id: 'eds', name: 'EDS 프로젝트', role: 'adm', roleLabel: '관리자' },
  { id: 'mkt', name: 'MKT 프로젝트', role: 'viewer', roleLabel: '조회자' }
];

// LDAP 그룹 정보 생성 헬퍼 함수
export const generateLDAPGroups = (user: AppUser): string[] => {
  return user.projects.map(project => `dataops/${project.id}/${project.role}`);
};

// 권한별 배지 색상 매핑
export const getRoleBadgeVariant = (role: 'dev' | 'adm' | 'viewer'): 'default' | 'destructive' | 'secondary' => {
  switch (role) {
    case 'adm':
      return 'destructive';
    case 'dev':
      return 'default';
    case 'viewer':
      return 'secondary';
    default:
      return 'default';
  }
};
