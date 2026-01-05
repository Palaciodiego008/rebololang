# Frontend Framework Support

ReboloLang now supports generating modern frontend applications with React, Svelte, or Vue, all powered by Bun.js and Vite for blazing-fast builds.

## 🚀 Quick Start

### Create a New App with Frontend

```bash
# React
rebolo new myapp --frontend react

# Svelte
rebolo new myapp --frontend svelte

# Vue
rebolo new myapp --frontend vue

# No frontend (default)
rebolo new myapp
```

### Short flag version:
```bash
rebolo new myapp -f react
rebolo new myapp -f svelte
rebolo new myapp -f vue
```

## 📁 Project Structure with Frontend

```
myapp/
├── frontend/                 # Frontend application
│   ├── src/
│   │   ├── App.tsx           # Main component (React)
│   │   ├── App.svelte        # Main component (Svelte)
│   │   ├── App.vue           # Main component (Vue)
│   │   ├── index.tsx/main.js # Entry point
│   │   ├── styles.css        # Styles
│   │   └── components/       # Your components
│   ├── package.json
│   ├── vite.config.js        # Vite build config
│   └── index.html            # HTML template
├── public/                   # Compiled frontend (auto-generated)
│   ├── index.html
│   └── assets/               # JS/CSS bundles
├── controllers/              # Go API controllers
├── models/                   # Go models
└── main.go                   # Go server (serves frontend + API)
```

## 🛠️ Development - Un Solo Comando

```bash
cd myapp
rebolo dev
```

¡Eso es todo! El comando `rebolo dev` automáticamente:
- ✅ Instala dependencias del frontend (si es necesario)
- ✅ Compila el frontend a `public/`
- ✅ Inicia el servidor Go en puerto 3000
- ✅ Vigila cambios en frontend y recompila automáticamente
- ✅ Vigila cambios en Go y reinicia el servidor

**Un solo terminal, un solo comando.**

### Lo que hace `rebolo dev`:

```
🎨 Frontend detected
📦 Installing frontend dependencies...
✅ Dependencies installed
⚡ Building frontend...
✅ Frontend built successfully
👀 Watching frontend for changes...
🔥 Starting Go server with hot reload...
🚀 Server running on http://localhost:3000
```

Visita `http://localhost:3000` - ¡Todo está listo!

## 🔄 Hot Reload Automático

- **Frontend**: Cambios en `.tsx`, `.vue`, `.svelte`, `.css` → Recompila automáticamente
- **Backend**: Cambios en `.go` → Reinicia el servidor automáticamente

## 🎨 Framework Features

### React + TypeScript
- Full TypeScript support
- React 18 with hooks
- Vite for ultra-fast builds
- Bun as package manager

### Svelte + Vite
- Svelte 4
- Vite for lightning-fast HMR
- Component-based architecture

### Vue 3 + Vite
- Vue 3 Composition API
- Vite for development
- Single File Components
- TypeScript ready

## 🔌 API Integration

El frontend se comunica con el backend a través de `/api/*`:

```javascript
// React/Svelte/Vue - Llamar API desde el frontend
async function getPosts() {
  const response = await fetch('/api/posts');
  const posts = await response.json();
  return posts;
}

async function createPost(data) {
  const response = await fetch('/api/posts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  });
  return response.json();
}
```

### Backend (Go)

```go
// main.go
app.GET("/api/posts", controllers.ListPosts)
app.POST("/api/posts", controllers.CreatePost)

// controllers/post_controller.go
func ListPosts(w http.ResponseWriter, r *http.Request) {
    posts := []Post{ /* ... */ }
    json.NewEncoder(w).Encode(posts)
}
```

## 📦 Why Bun.js?

- **⚡ 3x faster** than npm/yarn
- **🔥 Built-in bundler** - no extra config needed
- **📦 All-in-one** - package manager, bundler, and runtime
- **🚀 TypeScript native** - no transpilation needed
- **💚 Drop-in replacement** for Node.js

## 🚀 Build for Production

```bash
cd myapp
rebolo build  # Compiles frontend + Go binary
```

O manualmente:

```bash
# 1. Build frontend
cd frontend
bun run build

# 2. Build Go
cd ..
go build -o myapp

# 3. Deploy (copy binary + public/)
./myapp
```

## 🎯 Examples

### React Example
```tsx
import { useState, useEffect } from 'react';

function PostsList() {
  const [posts, setPosts] = useState([]);

  useEffect(() => {
    fetch('/api/posts')
      .then(res => res.json())
      .then(setPosts);
  }, []);

  return (
    <div>
      {posts.map(post => (
        <div key={post.id}>{post.title}</div>
      ))}
    </div>
  );
}
```

### Svelte Example
```svelte
<script>
  import { onMount } from 'svelte';
  let posts = [];

  onMount(async () => {
    const res = await fetch('/api/posts');
    posts = await res.json();
  });
</script>

{#each posts as post}
  <div>{post.title}</div>
{/each}
```

### Vue Example
```vue
<script setup>
import { ref, onMounted } from 'vue';

const posts = ref([]);

onMounted(async () => {
  const res = await fetch('/api/posts');
  posts.value = await res.json();
});
</script>

<template>
  <div v-for="post in posts" :key="post.id">
    {{ post.title }}
  </div>
</template>
```

## 🏗️ Architecture

```
┌────────────────────────────────────────┐
│   Browser (http://localhost:3000)     │
└────────────────┬───────────────────────┘
                 │
    ┌────────────▼────────────┐
    │   Go Server (:3000)     │
    ├─────────────────────────┤
    │  /              → SPA   │
    │  /assets/*      → Files │
    │  /api/*         → API   │
    └─────────────────────────┘
```

**Un solo servidor para todo.** Arquitectura monolítica moderna.

---

**Built with ❤️ for modern full-stack development**

