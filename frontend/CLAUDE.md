# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **uni-app** educational platform called "paperPlay" built with Vue 2, designed for academic exploration and learning. The app provides features like user authentication, daily challenges, goal tracking, interactive maps, and quiz functionality.

## Development Environment

### Prerequisites
- **HBuilderX** or **uni-app CLI** for development
- **Node.js** for package management
- **WeChat Developer Tools** or platform-specific dev tools for testing

### Key Commands

#### Development & Build
```bash
# Run in HBuilderX (recommended)
# Click "运行" button or use Ctrl+R

# CLI commands (if using uni-app CLI)
npm run dev:%PLATFORM%     # Development server
npm run build:%PLATFORM%   # Production build
```

#### Platform Targets
Replace `%PLATFORM%` with:
- `h5` - Web
- `mp-weixin` - WeChat Mini Program
- `app-plus` - App
- `mp-alipay` - Alipay Mini Program
- `mp-baidu` - Baidu Smart Program

#### Testing
```bash
# No formal test setup - manual testing through platform dev tools
# Check manifest.json for platform-specific configurations
```

## Architecture Overview

### Technology Stack
- **Frontend**: Vue 2 + uni-app framework
- **Styling**: SCSS (uni.scss)
- **API**: Custom HTTP client in `utils/request.js`
- **3rd Party Libraries**: 
  - Leaflet (maps)
  - PixiJS (graphics)
  - RoughJS (sketchy graphics)

### Directory Structure
```
src/
├── api/           # API service modules (auth, papers, questions, etc.)
├── pages/         # Page components (index, map, daily, goal, profile, quiz)
├── static/        # Static assets (images, fonts, SVGs)
├── utils/         # Utilities (request.js for API calls, web3.js)
├── hybrid/html/   # Web-specific assets (Leaflet CSS/JS, map.html)
└── App.vue        # Root application component
```

### Pages & Features
- **`pages/index/`** - Authentication (login/register)
- **`pages/map/`** - Interactive academic map using Leaflet
- **`pages/daily/`** - Daily challenges/recommendations
- **`pages/goal/`** - Goal setting and tracking
- **`pages/profile/`** - User profile management
- **`pages/quiz/`** - Quiz functionality

### API Configuration
- **Base URL**: `https://paperplay.zsh.cool`
- **Authentication**: JWT token stored in `uni.getStorageSync('token')`
- **Client**: Custom request wrapper in `utils/request.js`

### Key Files
- **manifest.json**: Platform configurations and permissions
- **pages.json**: Page routing and navigation configuration
- **uni.scss**: Global styling variables
- **App.vue**: Application lifecycle hooks

### Platform-Specific Notes
- **Vue 2** is configured (manifest.json:71)
- **Custom navigation** is used (navigationStyle: "custom" in pages.json)
- **Cross-platform compatibility** via uni-app's conditional compilation

### Development Tips
- Use HBuilderX for optimal uni-app development experience
- All pages use custom navigation bars (no system navigation)
- Images and assets are stored in `/static/` directory
- API endpoints are organized by feature in `/api/` directory