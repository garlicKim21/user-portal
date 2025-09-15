import { useState, useEffect, useRef } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { X, Terminal, Loader2, AlertCircle } from 'lucide-react';
import { launchTerminal, deleteTerminal } from '../services/terminalService';

interface WebTerminalProps {
  isOpen: boolean;
  onClose: () => void;
  user?: any;
}


export function WebTerminal({ isOpen, onClose, user }: WebTerminalProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [terminalUrl, setTerminalUrl] = useState<string | null>(null);
  const [resourceId, setResourceId] = useState<string | null>(null);
  const iframeRef = useRef<HTMLIFrameElement>(null);

  // 터미널 시작 함수
  const handleLaunchTerminal = async () => {
    setIsLoading(true);
    setError(null);
    setTerminalUrl(null);
    setResourceId(null);

    try {
      const result = await launchTerminal();
      setTerminalUrl(result.url);
      setResourceId(result.resourceId);
    } catch (err) {
      console.error('Terminal launch error:', err);
      setError(err instanceof Error ? err.message : 'Failed to launch terminal');
    } finally {
      setIsLoading(false);
    }
  };

  // 컴포넌트가 열릴 때 터미널 자동 시작
  useEffect(() => {
    if (isOpen && !terminalUrl && !isLoading) {
      handleLaunchTerminal();
    }
  }, [isOpen]);

  // 터미널 정리 함수
  const cleanupTerminal = async () => {
    if (resourceId) {
      try {
        await deleteTerminal(resourceId);
      } catch (err) {
        console.error('Failed to cleanup terminal:', err);
      }
    }
  };

  // 팝업이 닫힐 때 터미널 정리
  useEffect(() => {
    if (!isOpen && resourceId) {
      cleanupTerminal();
    }
  }, [isOpen, resourceId]);

  // ESC 키로 팝업 닫기
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleKeyDown);
      return () => document.removeEventListener('keydown', handleKeyDown);
    }
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div 
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          onClose();
        }
      }}
    >
      <Card className="w-[95vw] h-[95vh] max-w-7xl max-h-[90vh] flex flex-col shadow-2xl border-0 bg-card">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4 border-b bg-muted/50">
          <div className="flex items-center space-x-2">
            <Terminal className="h-6 w-6 text-primary" />
            <CardTitle className="text-xl">Secure Web Terminal</CardTitle>
            {user && (
              <span className="text-sm text-muted-foreground">
                ({user.preferred_username || user.name || 'User'})
              </span>
            )}
          </div>
          <div className="flex items-center space-x-2">
            {terminalUrl && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleLaunchTerminal}
                disabled={isLoading}
                className="hover:bg-primary/10"
              >
                {isLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  '새로고침'
                )}
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={onClose}
              className="hover:bg-destructive/10 hover:text-destructive"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>
        
        <CardContent className="flex-1 p-0 bg-background">
          {isLoading && (
            <div className="flex items-center justify-center h-full bg-muted/20">
              <div className="text-center space-y-4">
                <div className="relative">
                  <Loader2 className="h-12 w-12 animate-spin mx-auto text-primary" />
                  <Terminal className="h-6 w-6 absolute top-3 left-1/2 transform -translate-x-1/2 text-primary/60" />
                </div>
                <div>
                  <p className="text-lg font-medium">터미널을 시작하는 중...</p>
                  <p className="text-sm text-muted-foreground">잠시만 기다려주세요</p>
                </div>
              </div>
            </div>
          )}

          {error && (
            <div className="flex items-center justify-center h-full bg-muted/20">
              <div className="text-center space-y-4 max-w-md">
                <AlertCircle className="h-12 w-12 text-destructive mx-auto" />
                <div>
                  <p className="text-lg font-medium text-destructive mb-2">터미널 시작 실패</p>
                  <p className="text-sm text-muted-foreground mb-4">{error}</p>
                </div>
                <Button onClick={handleLaunchTerminal} disabled={isLoading} size="lg">
                  {isLoading ? (
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  ) : null}
                  다시 시도
                </Button>
              </div>
            </div>
          )}

          {terminalUrl && !isLoading && !error && (
            <div className="h-full relative">
              <iframe
                ref={iframeRef}
                src={terminalUrl}
                className="w-full h-full border-0 bg-background"
                title="Web Terminal"
                sandbox="allow-same-origin allow-scripts allow-forms allow-popups allow-popups-to-escape-sandbox"
              />
              <div className="absolute top-2 right-2 text-xs text-muted-foreground bg-background/80 px-2 py-1 rounded">
                Secure Terminal
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
