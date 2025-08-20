# WebSocket Examples

This directory contains 6 complete WebSocket example applications demonstrating different real-time features.

## Available Examples

### 1. Chat Application (`chat.html`)
**URL**: http://localhost:8100/static/chat.html

- Multi-room chat system
- Real-time messaging between users
- User presence indicators  
- Room switching capabilities
- Message history per room

**Features**:
- Join different chat rooms
- See who's online in each room
- Send and receive messages instantly
- Clean, responsive interface

### 2. Collaborative Drawing (`draw.html`)
**URL**: http://localhost:8100/static/draw.html

- Real-time collaborative drawing board
- Multi-user cursor tracking
- Synchronized drawing actions
- Color and brush size selection

**Features**:
- Draw collaboratively with other users
- See other users' cursors in real-time
- Synchronized canvas clearing
- Color and brush size controls

### 3. Kanban Board (`kanban.html`)
**URL**: http://localhost:8100/static/kanban.html

- Collaborative task management
- Real-time card movements between columns
- Task creation and editing
- Status updates across all users

**Features**:
- Create, edit, and move tasks
- Real-time updates when others make changes
- Multiple kanban columns (Todo, In Progress, Done)
- Collaborative project management

### 4. Code Editor (`editor.html`)
**URL**: http://localhost:8100/static/editor.html

- Collaborative code editing
- Real-time text synchronization
- Multiple programming languages
- Syntax highlighting

**Features**:
- Edit code collaboratively
- See changes from other users instantly
- Syntax highlighting for various languages
- Real-time cursor positions

### 5. System Monitor (`monitor.html`)
**URL**: http://localhost:8100/static/monitor.html

- Real-time system metrics display
- Performance monitoring dashboard
- Live data updates
- Interactive charts and graphs

**Features**:
- Monitor system performance metrics
- Real-time data visualization
- Multiple chart types
- Automatic updates

### 6. Spreadsheet (`spreadsheet.html`)
**URL**: http://localhost:8100/static/spreadsheet.html

- Collaborative spreadsheet editing
- Cell-level synchronization
- Real-time formula calculations
- Multi-user editing support

**Features**:
- Edit cells collaboratively
- Real-time formula updates
- See other users' selections
- Synchronized data entry

## Quick Start

1. **Start the Base Framework server**:
   ```bash
   base start
   ```

2. **Ensure WebSocket is enabled** (default):
   ```bash
   # Check .env file
   WS_ENABLED=true
   ```

3. **Access examples**:
   - Browse to http://localhost:8100/static/
   - Click on any example to open it
   - Enter a nickname and connect
   - Open multiple browser tabs to test collaboration

## WebSocket Connection Details

All examples connect to: `ws://localhost:8100/api/ws`

**Required Parameters**:
- `id`: Unique client identifier  
- `nickname`: Display name for the user
- `room`: Room name (specific to each example)

**Example Connection**:
```javascript
const socket = new WebSocket('ws://localhost:8100/api/ws?id=user123&nickname=John&room=general');
```

## Customization

Each example is self-contained and can be customized:

1. **HTML Structure**: Modify the UI layout and styling
2. **JavaScript Logic**: Add new features or modify behavior  
3. **WebSocket Messages**: Extend with custom message types
4. **Styling**: Update CSS for different themes

## Message Types

The examples use various WebSocket message types:

- `chat`: Text messages between users
- `system`: System notifications (join/leave)
- `users_update`: Updated list of connected users
- `draw`: Drawing coordinates and styles
- `kanban_update`: Task movements and changes
- `code_update`: Code editor changes
- `cursor_move`: Real-time cursor positions

## Browser Compatibility

All examples work with modern browsers that support:
- WebSocket API
- ES6 JavaScript features
- HTML5 Canvas (for drawing example)
- CSS Grid/Flexbox

## Development Tips

1. **Testing**: Open multiple browser tabs to simulate multiple users
2. **Debugging**: Use browser developer tools to inspect WebSocket messages
3. **Network**: Check Network tab for WebSocket connection status
4. **Console**: Look for JavaScript errors in console if features don't work

## Integration

These examples can be integrated into your Base Framework application:

1. **Copy relevant code** from examples into your application
2. **Extend message types** for your specific use case
3. **Add authentication** by requiring JWT tokens
4. **Persist data** using Base Framework's database integration

For complete implementation details, see `WEBSOCKET_IMPLEMENTATION.md` in the project root.