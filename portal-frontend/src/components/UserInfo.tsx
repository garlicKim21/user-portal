import { useState } from 'react';
import { User as UserIcon, LogOut, ChevronDown } from 'lucide-react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Separator } from './ui/separator';
import { AppUser, generateLDAPGroups } from '../types/user';

interface UserInfoProps {
  user: AppUser;
  onLogout: () => void;
}

export function UserInfo({ user, onLogout }: UserInfoProps) {
  const [open, setOpen] = useState(false);
  const ldapGroups = generateLDAPGroups(user);

  console.log('UserInfo rendering with user:', user); // 디버깅용

  return (
    <div className="flex items-center gap-2">
      <Button 
        variant="ghost" 
        className="flex items-center gap-2"
        onClick={() => setOpen(!open)}
      >
        <UserIcon className="h-4 w-4" />
        <span>
          {user.family_name && user.given_name 
            ? `${user.family_name}${user.given_name}` 
            : user.name || 'Unknown User'
          }
        </span>
        <ChevronDown className="h-3 w-3 opacity-50" />
      </Button>
      
      {open && (
        <div className="absolute right-4 top-16 w-80 bg-white border rounded-lg shadow-lg z-50">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-lg flex items-center gap-2">
                <UserIcon className="h-5 w-5" />
                사용자 정보
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* 기본 사용자 정보 (단순화) */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">사용자 ID</span>
                  <span className="text-sm text-muted-foreground">{user.preferred_username || '정보 없음'}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">사용자 이름</span>
                  <span className="text-sm text-muted-foreground">
                    {user.family_name && user.given_name 
                      ? `${user.family_name}${user.given_name}` 
                      : user.name || '정보 없음'
                    }
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">이메일</span>
                  <span className="text-sm text-muted-foreground">{user.email || '정보 없음'}</span>
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
        </div>
      )}
    </div>
  );
}
