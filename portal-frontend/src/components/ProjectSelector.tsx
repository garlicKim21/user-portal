import { useState } from 'react';
import { Check, ChevronsUpDown, Building2 } from 'lucide-react';
import { cn } from './ui/utils';
import { Button } from './ui/button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from './ui/command';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';
import { UserProject } from '../types/user';

interface ProjectSelectorProps {
  projects: UserProject[];
  currentProject: UserProject | null;
  onProjectChange: (project: UserProject) => void;
}

export function ProjectSelector({ projects, currentProject, onProjectChange }: ProjectSelectorProps) {
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
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
      </PopoverTrigger>
      <PopoverContent className="w-[280px] p-0">
        <Command>
          <CommandInput placeholder="프로젝트 검색..." />
          <CommandList>
            <CommandEmpty>프로젝트를 찾을 수 없습니다.</CommandEmpty>
            <CommandGroup>
              {projects.map((project) => (
                <CommandItem
                  key={project.id}
                  value={`${project.name} ${project.id}`}
                  onSelect={() => {
                    onProjectChange(project);
                    setOpen(false);
                  }}
                >
                  <div className="flex items-center justify-between w-full">
                    <div className="flex items-center gap-2">
                      <Building2 className="h-4 w-4" />
                      <span>{project.name}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-xs px-2 py-1 bg-gray-200 rounded">
                        {project.roleLabel}
                      </span>
                      <Check
                        className={cn(
                          "ml-auto h-4 w-4",
                          currentProject?.id === project.id ? "opacity-100" : "opacity-0"
                        )}
                      />
                    </div>
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
