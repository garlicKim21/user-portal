import {useEffect, useState} from "react";
import {useAuth} from "react-oidc-context";
import {backendAuthService} from "@/services/backendAuthService.ts";
import {verifyTerminalReady} from "@/services/terminalService.ts";
import {Spinner} from "@/components/ui/spinner.tsx";
import TerminalViewer from "@/components/ui/terminal-viewer.tsx";
import {Button} from "@/components/ui/button.tsx";

const TERM_URL_SESSION_KEY = 'TERM_URL_SESSION_KEY';

const Terminal = () => {
    const { user } = useAuth();
    const [isLoading, setIsLoading] = useState(true);
    const [statusText, setStatusText] = useState<string>("터미널을 준비하고 있습니다...");
    const [errorText, setErrorText] = useState<string| null>(null);
    const [terminalUrl, setTerminalUrl] = useState<string | null>(null);

    useEffect(() => {
        if (!user?.access_token) {
            return;
        }

        const validateAndSetUrl = async (url: string): Promise<boolean> => {
            const isReady = await verifyTerminalReady(url);
            if (isReady) {
                setTerminalUrl(url);
                sessionStorage.setItem(TERM_URL_SESSION_KEY, url);
                return true;
            }
            return false;
        };

        const getTerminalUrl = async () => {
            try {
                // 저장된 URL 확인
                const savedUrl = sessionStorage.getItem(TERM_URL_SESSION_KEY);
                if (savedUrl) {
                    setStatusText("연결을 확인하고 있습니다...");
                    const isValid = await validateAndSetUrl(savedUrl);
                    if (isValid) return;

                    // 유효하지 않으면 삭제
                    sessionStorage.removeItem(TERM_URL_SESSION_KEY);
                }

                // 새 URL 생성
                setStatusText("웹 터미널 관련 리소스를 생성하고 있습니다...");
                const result = await backendAuthService.launchWebConsole(user.access_token);
                const isValidUrl = await validateAndSetUrl(result.url);

                if (!isValidUrl) {
                    setErrorText("웹 터미널 연결 타임 아웃. 연결을 다시 시도하여 주세요.");
                }
            } catch (error) {
                setErrorText("터미널 생성 중 서버 오류가 발생했습니다. 탭을 닫고 다시 시도하여 주세요");
            } finally {
                setIsLoading(false);
            }
        };

        void getTerminalUrl();
    }, [user]);

    const handleRefresh = () => {
        window.location.reload();
    };

    const handleRetry = () => {
        sessionStorage.removeItem(TERM_URL_SESSION_KEY);
        window.location.reload();
    };

    if (terminalUrl) {
        return <TerminalViewer url={terminalUrl} />;
    }

    return (
        <div className="flex flex-col items-center justify-center h-screen gap-4">
            {isLoading && <Spinner />}
            {!errorText && (
                <div className="text-gray-600">{statusText}</div>
            )}

            {errorText && (
                <>
                    <div style={{ fontWeight: 600 }} className="text-red-500">{errorText}</div>
                    <div className="flex gap-3">
                        <Button onClick={handleRefresh} variant="outline">
                            연결 재시도
                        </Button>
                        <Button onClick={handleRetry}>
                            WebTerminal 재성성
                        </Button>
                    </div>
                </>
            )}
        </div>
    );
}

export default Terminal;