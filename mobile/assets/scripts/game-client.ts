// Game Client for Cocos Creator
// Handles WebSocket connection and game state synchronization

const { ccclass, property } = cc._decorator;

@ccclass('GameClient')
export default class GameClient extends cc.Component {
    @property
    private serverUrl: string = 'ws://localhost:3002';

    @property
    private tableId: string = '';

    private socket: WebSocket = null;
    private playerId: string = '';
    private reconnectAttempts: number = 0;
    private maxReconnectAttempts: number = 5;
    private reconnectDelay: number = 1000;

    onLoad() {
        console.log('GameClient initialized');
    }

    connect(tableId: string, playerId: string) {
        this.tableId = tableId;
        this.playerId = playerId;
        this.serverUrl = `${this.serverUrl}/${tableId}`;
        this.establishConnection();
    }

    private establishConnection() {
        try {
            this.socket = new WebSocket(this.serverUrl);

            this.socket.onopen = () => {
                console.log('Connected to game server');
                this.reconnectAttempts = 0;
                this.joinTable();
            };

            this.socket.onmessage = (event) => {
                this.handleMessage(JSON.parse(event.data));
            };

            this.socket.onclose = (event) => {
                console.log('Disconnected from game server', event.code, event.reason);
                this.attemptReconnect();
            };

            this.socket.onerror = (error) => {
                console.error('WebSocket error:', error);
            };
        } catch (error) {
            console.error('Failed to establish connection:', error);
            this.attemptReconnect();
        }
    }

    private attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
            setTimeout(() => {
                this.establishConnection();
            }, this.reconnectDelay * this.reconnectAttempts);
        } else {
            console.error('Max reconnect attempts reached');
        }
    }

    private joinTable() {
        this.send({
            type: 'join',
            player_id: this.playerId,
            player_name: 'Player', // Would come from user profile
            chips: 1000
        });
    }

    private handleMessage(message: any) {
        switch (message.type) {
            case 'joined':
                console.log('Successfully joined table');
                this.onGameStateUpdate(message.state);
                break;
            case 'state_update':
                this.onGameStateUpdate(message);
                break;
            case 'player_joined':
                console.log('Player joined:', message.player_id);
                break;
            case 'player_left':
                console.log('Player left:', message.player_id);
                break;
            case 'pot_won':
                console.log('Pot won by:', message.player_id, 'amount:', message.amount);
                break;
            case 'error':
                console.error('Server error:', message.message);
                break;
            default:
                console.log('Unknown message type:', message.type);
        }
    }

    private onGameStateUpdate(state: any) {
        // Emit event for other components to handle
        cc.game.emit('game_state_update', state);
    }

    sendAction(action: string, amount?: number) {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error('Not connected to server');
            return;
        }

        this.send({
            type: 'action',
            player_id: this.playerId,
            action: action,
            amount: amount || 0
        });
    }

    leaveTable() {
        if (this.socket) {
            this.send({
                type: 'leave',
                player_id: this.playerId
            });
            this.socket.close();
        }
    }

    private send(data: any) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(data));
        }
    }

    onDestroy() {
        this.leaveTable();
    }
}
