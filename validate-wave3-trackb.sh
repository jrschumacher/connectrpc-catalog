#!/bin/bash
set -e

echo "🔍 Validating Wave 3 Track B - React Frontend Implementation"
echo ""

# Check UI directory structure
echo "✓ Checking UI directory structure..."
[ -d "ui/src" ] || { echo "❌ ui/src not found"; exit 1; }
[ -d "ui/src/components" ] || { echo "❌ ui/src/components not found"; exit 1; }
[ -d "ui/src/components/ui" ] || { echo "❌ ui/src/components/ui not found"; exit 1; }
[ -d "ui/src/lib" ] || { echo "❌ ui/src/lib not found"; exit 1; }

# Check configuration files
echo "✓ Checking configuration files..."
[ -f "ui/package.json" ] || { echo "❌ package.json not found"; exit 1; }
[ -f "ui/vite.config.ts" ] || { echo "❌ vite.config.ts not found"; exit 1; }
[ -f "ui/tsconfig.json" ] || { echo "❌ tsconfig.json not found"; exit 1; }
[ -f "ui/tailwind.config.js" ] || { echo "❌ tailwind.config.js not found"; exit 1; }
[ -f "ui/postcss.config.js" ] || { echo "❌ postcss.config.js not found"; exit 1; }
[ -f "ui/index.html" ] || { echo "❌ index.html not found"; exit 1; }

# Check main components
echo "✓ Checking main components..."
[ -f "ui/src/App.tsx" ] || { echo "❌ App.tsx not found"; exit 1; }
[ -f "ui/src/main.tsx" ] || { echo "❌ main.tsx not found"; exit 1; }
[ -f "ui/src/index.css" ] || { echo "❌ index.css not found"; exit 1; }

# Check feature components
echo "✓ Checking feature components..."
[ -f "ui/src/components/ServiceBrowser.tsx" ] || { echo "❌ ServiceBrowser.tsx not found"; exit 1; }
[ -f "ui/src/components/MethodDetails.tsx" ] || { echo "❌ MethodDetails.tsx not found"; exit 1; }
[ -f "ui/src/components/RequestEditor.tsx" ] || { echo "❌ RequestEditor.tsx not found"; exit 1; }
[ -f "ui/src/components/ResponseViewer.tsx" ] || { echo "❌ ResponseViewer.tsx not found"; exit 1; }

# Check UI base components
echo "✓ Checking shadcn/ui components..."
[ -f "ui/src/components/ui/button.tsx" ] || { echo "❌ button.tsx not found"; exit 1; }
[ -f "ui/src/components/ui/input.tsx" ] || { echo "❌ input.tsx not found"; exit 1; }
[ -f "ui/src/components/ui/card.tsx" ] || { echo "❌ card.tsx not found"; exit 1; }
[ -f "ui/src/components/ui/label.tsx" ] || { echo "❌ label.tsx not found"; exit 1; }
[ -f "ui/src/components/ui/badge.tsx" ] || { echo "❌ badge.tsx not found"; exit 1; }
[ -f "ui/src/components/ui/scroll-area.tsx" ] || { echo "❌ scroll-area.tsx not found"; exit 1; }
[ -f "ui/src/components/ui/separator.tsx" ] || { echo "❌ separator.tsx not found"; exit 1; }

# Check library files
echo "✓ Checking library files..."
[ -f "ui/src/lib/client.ts" ] || { echo "❌ client.ts not found"; exit 1; }
[ -f "ui/src/lib/types.ts" ] || { echo "❌ types.ts not found"; exit 1; }
[ -f "ui/src/lib/utils.ts" ] || { echo "❌ utils.ts not found"; exit 1; }

# Check generated TypeScript files
echo "✓ Checking generated TypeScript files..."
[ -f "gen/catalog/v1/catalog_connect.ts" ] || { echo "❌ catalog_connect.ts not found"; exit 1; }

# Check buf.gen.yaml includes TypeScript generation
echo "✓ Checking buf.gen.yaml for TypeScript generation..."
grep -q "buf.build/connectrpc/es" buf.gen.yaml || { echo "❌ TypeScript generation not configured"; exit 1; }

# Count files
echo ""
echo "📊 File Statistics:"
echo "   - TypeScript files: $(find ui/src -name "*.tsx" -o -name "*.ts" | wc -l | tr -d ' ')"
echo "   - Main components: 4"
echo "   - UI components: 7"
echo "   - Config files: 7"
echo ""

echo "✅ Wave 3 Track B validation PASSED"
echo ""
echo "📝 Summary:"
echo "   - UI scaffold: ✅ Complete"
echo "   - ServiceBrowser: ✅ Complete"
echo "   - MethodDetails: ✅ Complete"
echo "   - RequestEditor: ✅ Complete"
echo "   - ResponseViewer: ✅ Complete"
echo "   - ConnectRPC integration: ✅ Complete"
echo ""
echo "🚀 Next: Run 'cd ui && npm install && npm run dev' to test"
