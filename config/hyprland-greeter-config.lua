-- SYSC-Greet Hyprland Lua config for greetd greeter session

hl.monitor({ output = "", mode = "preferred", position = "auto", scale = "auto" })

hl.config({
    animations = { enabled = false },
    decoration = {
        rounding = 0,
        blur = { enabled = false },
        shadow = { enabled = false },
    },
    general = { gaps_in = 0, gaps_out = 0, border_size = 0 },
    input = { kb_layout = "us", repeat_delay = 400, repeat_rate = 40 },
    misc = { disable_hyprland_logo = true },
})

hl.window_rule({ match = { class = "kitty" }, rounding = 0 })

hl.on("hyprland.start", function()
    hl.exec_cmd("cd /var/lib/greeter && env HOME=/var/lib/greeter XDG_CACHE_HOME=/var/lib/greeter/.cache XDG_CONFIG_HOME=/var/lib/greeter/.config XDG_STATE_HOME=/var/lib/greeter/.local/state /usr/local/bin/sysc-greet --wallpaper-daemon")
    hl.exec_cmd("cd /var/lib/greeter && env HOME=/var/lib/greeter XDG_CACHE_HOME=/var/lib/greeter/.cache XDG_CONFIG_HOME=/var/lib/greeter/.config XDG_STATE_HOME=/var/lib/greeter/.local/state kitty --start-as=fullscreen --config=/etc/greetd/kitty.conf /usr/local/bin/sysc-greet && hyprctl dispatch 'hl.dsp.exit()'")
end)
