# ReboloLang CLI Commands 🚀

## Installation
```bash
curl -fsSL https://raw.githubusercontent.com/Palaciodiego008/rebololang/main/install.sh | bash
```

## Commands

### App Management
```bash
rebolo new myapp              # Create new application
rebolo dev                    # Start development server with hot reload
```

### Code Generation
```bash
rebolo generate resource posts title:string content:text published:bool
rebolo g resource users name:string email:string age:int    # shorthand
```

### Database Operations
```bash
rebolo db migrate             # Run database migrations
```

## Quick Start
```bash
# Create a blog app
rebolo new blog
cd blog

# Generate posts resource
rebolo g resource posts title:string content:text published:bool

# Start development server
rebolo dev
```

## Example Workflow
```bash
# 1. Create app
rebolo new ecommerce
cd ecommerce

# 2. Generate resources
rebolo g resource products name:string price:float description:text
rebolo g resource users name:string email:string
rebolo g resource orders user_id:int total:float

# 3. Configure database (edit config.yml)
# database:
#   url: postgres://localhost/ecommerce_development

# 4. Run migrations
rebolo db migrate

# 5. Start development
rebolo dev
```

Your app will be running at `http://localhost:3000` with:
- ✅ Hot reload for Go and frontend
- ✅ Complete CRUD interfaces
- ✅ Beautiful styled forms
- ✅ Database integration ready
- ✅ Colombian pride! 🇨🇴

**¡Vamos Rebolo!** 🚀
