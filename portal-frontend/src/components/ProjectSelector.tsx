import { useState } from 'react';
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

  console.log('ProjectSelector rendering with projects:', projects); // 디버깅용

  return (
    <div className="relative">
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
        <div className="absolute top-full left-0 mt-1 w-[280px] bg-white border rounded-lg shadow-lg z-50">
          <div className="p-2 space-y-1">
            {projects.map((project) => (
              <div
                key={project.id}
                onClick={() => {
                  onProjectChange(project);
                  setOpen(false);
                }}
                className={`p-2 rounded cursor-pointer hover:bg-gray-100 ${
                  currentProject?.id === project.id ? 'bg-blue-50 border border-blue-200' : ''
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Building2 className="h-4 w-4" />
                    <span className="font-medium">{project.name}</span>
                  </div>
                  <span className="text-xs px-2 py-1 bg-gray-200 rounded">
                    {project.roleLabel}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
