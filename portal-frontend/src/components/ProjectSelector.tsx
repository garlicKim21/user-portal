import { useState, useEffect, useRef } from 'react';
import { ChevronsUpDown, Building2 } from 'lucide-react';
import { Button } from './ui/button';
import { UserProject } from '../types/user';

interface ProjectSelectorProps {
  projects: UserProject[];
  currentProject: UserProject | null;
  onProjectChange: (project: UserProject) => void;
}

export function ProjectSelector({ projects, currentProject, onProjectChange }: ProjectSelectorProps) {
  const [open, setOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  console.log('ProjectSelector rendering with projects:', projects); // 디버깅용

  // 외부 클릭 시 드롭다운 닫기
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    }

    if (open) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [open]);

  return (
    <div className="relative" ref={dropdownRef}>
      <Button
        variant="outline"
        onClick={() => setOpen(!open)}
        className="w-[280px] justify-between"
      >
        <div className="flex items-center gap-2">
          <Building2 className="h-4 w-4" />
          {currentProject ? (
            <div className="flex items-center gap-2">
              <span>{currentProject.name}</span>
              <span className="text-xs px-2 py-1 bg-gray-200 rounded">
                {currentProject.roleLabel}
              </span>
            </div>
          ) : (
            "프로젝트를 선택하세요..."
          )}
        </div>
        <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
      </Button>
      
      {open && (
        <>
          {/* 배경 오버레이 */}
          <div className="fixed inset-0 bg-black/20 z-40" onClick={() => setOpen(false)} />
          
          {/* 드롭다운 메뉴 */}
          <div className="absolute top-full left-0 mt-1 w-[280px] bg-white border-2 border-gray-400 rounded-lg shadow-2xl z-50" style={{backgroundColor: '#ffffff', opacity: 1}}>
            <div className="p-2 space-y-1" style={{backgroundColor: '#ffffff'}}>
              {projects.map((project) => (
                <div
                  key={project.id}
                  onClick={() => {
                    onProjectChange(project);
                    setOpen(false);
                  }}
                  className={`p-3 rounded cursor-pointer transition-colors border ${
                    currentProject?.id === project.id ? 'bg-blue-100 border-blue-400' : 'bg-white border-transparent hover:bg-gray-100 hover:border-gray-300'
                  }`}
                  style={{backgroundColor: currentProject?.id === project.id ? '#dbeafe' : '#ffffff'}}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Building2 className="h-4 w-4" />
                      <span className="font-medium">{project.name}</span>
                    </div>
                    <span className="text-xs px-2 py-1 bg-gray-200 rounded font-medium">
                      {project.roleLabel}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
