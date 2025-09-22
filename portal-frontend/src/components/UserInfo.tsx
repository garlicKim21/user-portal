import { useState } from 'react';
import { User, LogOut, ChevronDown } from 'lucide-react';
import { Button } from './ui/button';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';
import { Badge } from './ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Separator } from './ui/separator';
import { AppUser, generateLDAPGroups, getRoleBadgeVariant } from '../types/user';

interface UserInfoProps {
  user: AppUser;
  onLogout: () => void;
}

export function UserInfo({ user, onLogout }: UserInfoProps) {
  const [open, setOpen] = useState(false);
  const ldapGroups = generateLDAPGroups(user);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" className="flex items-center gap-2">
          <User className="h-4 w-4" />
          <span>{user.name}</span>
          <ChevronDown className="h-3 w-3 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80" align="end">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-lg flex items-center gap-2">
              <User className="h-5 w-5" />
              사용자 정보
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* 기본 사용자 정보 */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">이름</span>
                <span className="text-sm text-muted-foreground">{user.name}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">사용자명</span>
                <span className="text-sm text-muted-foreground">{user.preferred_username}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">이메일</span>
                <span className="text-sm text-muted-foreground">{user.email || '정보 없음'}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">이메일 인증</span>
                <span className={`text-sm ${user.email_verified ? 'text-green-600' : 'text-red-600'}`}>
                  {user.email_verified ? '인증됨' : '미인증'}
                </span>
              </div>
            </div>

            <Separator />

            {/* 소속 프로젝트 목록 */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium">소속 프로젝트</h4>
              <div className="space-y-2">
                {user.projects.map((project) => (
                  <div key={project.id} className="flex items-center justify-between p-2 rounded-md bg-muted/50">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium">{project.name}</span>
                    </div>
                    <Badge variant={getRoleBadgeVariant(project.role)} className="text-xs">
                      {project.roleLabel}
                    </Badge>
                  </div>
                ))}
              </div>
            </div>

            <Separator />

            {/* LDAP 그룹 정보 */}
            <div className="space-y-3">
              <h4 className="text-sm font-medium">LDAP 그룹</h4>
              <div className="space-y-1">
                {ldapGroups.map((group, index) => (
                  <div key={index} className="text-xs font-mono text-muted-foreground bg-muted/30 p-2 rounded">
                    {group}
                  </div>
                ))}
              </div>
            </div>

            <Separator />

            {/* 로그아웃 버튼 */}
            <Button 
              variant="destructive" 
              onClick={() => {
                onLogout();
                setOpen(false);
              }}
              className="w-full"
            >
              <LogOut className="h-4 w-4 mr-2" />
              로그아웃
            </Button>
          </CardContent>
        </Card>
      </PopoverContent>
    </Popover>
  );
}
