// components/TerminalViewer.tsx
import React from 'react';

interface TerminalViewerProps {
    url: string;
    className?: string;
}

const TerminalViewer = ({ url, className = '' }: TerminalViewerProps) => {
    const handleCloseTab = () => {
        window.close();
    };

    const handleIframeError = (e: React.SyntheticEvent<HTMLIFrameElement>) => {
        console.error('Iframe loading error:', e);
    };

    return (
        <div className={`flex flex-col h-screen bg-background text-foreground font-korean ${className}`}>
            {/* 터미널 iframe */}
            <div className="flex-1 relative">
                <iframe
                    src={url}
                    className="w-full h-full border-0"
                    title="터미널"
                    onError={handleIframeError}
                    allow="clipboard-read; clipboard-write; fullscreen"
                    loading="lazy"
                />
            </div>

            {/* 하단 버튼 */}
            <div className="px-4 py-3 border-t border-border bg-muted/30">
                <div className="flex justify-end pr-2">
                    <button
                        onClick={handleCloseTab}
                        className="px-4 py-2 text-sm font-medium text-white bg-destructive hover:bg-destructive/90 focus:ring-2 focus:ring-ring focus:ring-offset-2 rounded-md transition-all duration-200 shadow-sm hover:shadow-md focus:outline-hidden"
                        type="button"
                    >
                        Terminal 종료
                    </button>
                </div>
            </div>
        </div>
    );
};

export default TerminalViewer;