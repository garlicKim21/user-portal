import { useState } from 'react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';

interface LoginPageProps {
  onLogin: () => void;
}

export function LoginPage({ onLogin }: LoginPageProps) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (username && password) {
      setIsLoading(true);
      
      // 실제로는 API 호출이 필요하지만, 여기서는 시뮬레이션
      try {
        await new Promise(resolve => setTimeout(resolve, 1500)); // 1.5초 로딩 시뮬레이션
        onLogin();
      } catch (error) {
        console.error('로그인 실패:', error);
      } finally {
        setIsLoading(false);
      }
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/30">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <div className="flex justify-center mb-4">
            <img 
              src="/skhynix.png" 
              alt="SK하이닉스 로고" 
              className="h-12 w-auto"
            />
          </div>
          <CardTitle className="text-center login-title">빅데이터 분석 플랫폼 사용자 포털</CardTitle>
          <CardDescription className="text-center login-description">
            개발자 및 빅데이터 분석가 전용 포털에 로그인하세요
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="username" className="login-label">사용자명</Label>
              <Input
                id="username"
                type="text"
                placeholder="사용자명을 입력하세요"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="login-input"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password" className="login-label">비밀번호</Label>
              <Input
                id="password"
                type="password"
                placeholder="비밀번호를 입력하세요"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="login-input"
                required
              />
            </div>
            <Button 
              type="submit" 
              className={`w-full login-button ${isLoading ? 'loading' : ''}`}
              disabled={isLoading || !username || !password}
            >
              {isLoading ? '로그인 중...' : '로그인'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}