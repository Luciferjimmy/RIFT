// Reactive state store with pub/sub
export const store = {
    state: {
        offlineMode: false,
        games: [],
        downloads: [],
        activeDownload: null,
        systemInfo: {},
        settings: {},
        runningGame: null,
        epicAccount: null,
        steamLoading: false,
        authStatus: {
            epicConnected: false,
            epicUsername: "",
            steamConnected: false,
            steamUsername: "",
            steamInstalled: false
        },
        profile: {
            name: "Abhinaw",
            alias: "Abhinaw",
            username: "abhinaw",
            avatar: "radial-gradient(circle at 30% 30%, #4a5d5e, #1a2022)",
            rank: 12,
            xp: 4500,
            xpNext: 5000,
            yearsOnRift: 1,
            achievements: {
                "collector": 2, // 2 out of 3 stars
                "veteran": 1    // 1 out of 3 stars
            }
        }
    },

    listeners: {},

    subscribe(key, callback) {
        if (!this.listeners[key]) this.listeners[key] = [];
        this.listeners[key].push(callback);
        callback(this.state[key]); // Immediate with current value
        return () => {
            this.listeners[key] = this.listeners[key].filter(cb => cb !== callback);
        };
    },

    update(key, value) {
        this.state[key] = value;
        if (this.listeners[key]) {
            this.listeners[key].forEach(cb => cb(value));
        }
    }
};
