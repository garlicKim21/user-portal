import { LoaderIcon } from "lucide-react"

function Spinner({ className, ...props }: React.ComponentProps<"svg">) {
  return (
    <LoaderIcon
      role="status"
      aria-label="Loading"
      className={`size-6 animate-spin ${className || ''}`}
      style={{
          animation: 'spin 1s linear infinite',
          ...props.style
      }}
      {...props}
    />
  )
}

export { Spinner }