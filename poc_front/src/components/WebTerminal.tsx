import { useState, useEffect } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { X, Terminal, Loader2, AlertCircle, ExternalLink } from 'lucide-react';
import { launchTerminal, deleteTerminal } from '../services/terminalService';

interface WebTerminalProps {
  isOpen: boolean;
  onClose: () => void;
  user?: any;
}

interface PopupWindow extends Window {
  closed: boolean;
}


export function WebTerminal({ isOpen, onClose }: WebTerminalProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [resourceId, setResourceId] = useState<string | null>(null);
  const [popupWindow, setPopupWindow] = useState<PopupWindow | null>(null);

  // 팝업 창 열기 함수
  const openTerminalPopup = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const result = await launchTerminal();
      setResourceId(result.resourceId);

      // 팝업 창 열기
      const popup = window.open(
        '',
        'web-terminal',
        'width=1200,height=800,scrollbars=yes,resizable=yes,toolbar=no,menubar=no,location=no,status=no'
      );

      if (!popup) {
        throw new Error('팝업이 차단되었습니다. 팝업 차단을 해제해주세요.');
      }

      setPopupWindow(popup);

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
                window.opener.postMessage({ type: 'terminal-closed' }, '*');
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
          setPopupWindow(null);
          onClose();
        }
      }, 1000);

    } catch (err) {
      console.error('Terminal launch error:', err);
      setError(err instanceof Error ? err.message : 'Failed to launch terminal');
    } finally {
      setIsLoading(false);
    }
  };


  // 컴포넌트가 열릴 때 팝업 창 열기
  useEffect(() => {
    if (isOpen && !popupWindow && !isLoading) {
      openTerminalPopup();
    }
  }, [isOpen]);

  // 팝업 창이 닫힐 때 정리
  useEffect(() => {
    if (!isOpen && popupWindow) {
      if (!popupWindow.closed) {
        popupWindow.close();
      }
      setPopupWindow(null);
    }
  }, [isOpen, popupWindow]);

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

  // 팝업 창에서 메시지 받기
  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.data.type === 'terminal-closed') {
        setPopupWindow(null);
        onClose();
      }
    };

    window.addEventListener('message', handleMessage);
    return () => window.removeEventListener('message', handleMessage);
  }, [onClose]);

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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <Card className="w-[500px] max-w-[90vw] flex flex-col shadow-2xl border-0 bg-card">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4 border-b bg-muted/50">
          <div className="flex items-center space-x-2">
            <Terminal className="h-6 w-6 text-primary" />
            <CardTitle className="text-xl">Secure Web Terminal</CardTitle>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="hover:bg-destructive/10 hover:text-destructive"
          >
            <X className="h-4 w-4" />
          </Button>
        </CardHeader>
        
        <CardContent className="p-6">
          {isLoading && (
            <div className="text-center space-y-4">
              <div className="relative">
                <Loader2 className="h-12 w-12 animate-spin mx-auto text-primary" />
                <Terminal className="h-6 w-6 absolute top-3 left-1/2 transform -translate-x-1/2 text-primary/60" />
              </div>
              <div>
                <p className="text-lg font-medium">터미널 팝업을 여는 중...</p>
                <p className="text-sm text-muted-foreground">잠시만 기다려주세요</p>
              </div>
            </div>
          )}

          {error && (
            <div className="text-center space-y-4">
              <AlertCircle className="h-12 w-12 text-destructive mx-auto" />
              <div>
                <p className="text-lg font-medium text-destructive mb-2">팝업 열기 실패</p>
                <p className="text-sm text-muted-foreground mb-4">{error}</p>
              </div>
              <Button onClick={openTerminalPopup} disabled={isLoading} size="lg">
                {isLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                ) : null}
                다시 시도
              </Button>
            </div>
          )}

          {popupWindow && !isLoading && !error && (
            <div className="text-center space-y-4">
              <div className="flex items-center justify-center space-x-2">
                <ExternalLink className="h-8 w-8 text-green-500" />
                <div>
                  <p className="text-lg font-medium">터미널 팝업이 열렸습니다</p>
                  <p className="text-sm text-muted-foreground">
                    새 창에서 터미널을 사용하실 수 있습니다
                  </p>
                </div>
              </div>
              <div className="space-y-2">
                <Button 
                  onClick={() => popupWindow.focus()} 
                  variant="outline" 
                  className="w-full"
                >
                  팝업 창으로 이동
                </Button>
                <Button 
                  onClick={onClose} 
                  variant="ghost" 
                  className="w-full"
                >
                  닫기
                </Button>
              </div>
            </div>
          )}

          {!popupWindow && !isLoading && !error && (
            <div className="text-center space-y-4">
              <Terminal className="h-16 w-16 mx-auto text-muted-foreground/50" />
              <div>
                <p className="text-lg font-medium text-muted-foreground">터미널 팝업 준비 중</p>
                <p className="text-sm text-muted-foreground/70">
                  팝업이 차단되었을 수 있습니다. 팝업 차단을 해제해주세요.
                </p>
              </div>
              <Button onClick={openTerminalPopup} disabled={isLoading} size="lg">
                {isLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                ) : null}
                터미널 팝업 열기
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
