# ConnectRPC Catalog UI

React frontend for the ConnectRPC Catalog service browser and testing tool.

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Styling
- **shadcn/ui** - UI component library
- **ConnectRPC** - gRPC-web client

## Development

```bash
# Install dependencies
npm install

# Start dev server (requires backend running on :8080)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Architecture

```
src/
├── components/
│   ├── ui/              # shadcn/ui base components
│   ├── ServiceBrowser   # Service/method tree navigation
│   ├── MethodDetails    # Method signature and schema display
│   ├── RequestEditor    # Form/JSON request builder
│   └── ResponseViewer   # Response display with syntax highlighting
├── lib/
│   ├── client.ts        # ConnectRPC client configuration
│   ├── types.ts         # TypeScript type definitions
│   └── utils.ts         # Utility functions
└── App.tsx              # Main application component
```

## Features

- Service and method browsing with search
- Dynamic request form generation from Protobuf schemas
- JSON request editor mode
- Real-time response viewing
- Error handling and loading states
- Responsive design

## Backend Integration

The UI connects to the CatalogService backend via ConnectRPC:

- `ListServices` - Load available services and methods
- `GetServiceSchema` - Fetch Protobuf schemas for request/response types
- `InvokeGRPC` - Proxy gRPC requests through the backend

## Build Output

Production builds are output to `dist/` and embedded into the Go binary via `embed.FS`.
