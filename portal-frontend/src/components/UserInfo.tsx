import { User as UserIcon } from 'lucide-react';
import { AppUser } from '../types/user';

interface UserInfoProps {
  user: AppUser;
}

export function UserInfo({ user }: UserInfoProps) {
  return (
    <div className="flex items-center gap-2">
      <UserIcon className="h-4 w-4" />
      <span className="text-sm font-medium">
        {user.family_name && user.given_name 
          ? `${user.family_name}${user.given_name}` 
          : user.name || 'Unknown User'
        }
      </span>
    </div>
  );
}
