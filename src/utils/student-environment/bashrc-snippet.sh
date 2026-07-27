export PATH=$PATH:/usr/sbin:/sbin
export EDITOR=nano
export VISUAL=nano

if [ -f "$HOME/.signed-complete" ]; then
    case "$(cat "$HOME/.signed-complete" 2>/dev/null)" in
        beagle)
            PS1="(Beagle-junior) \u@\h:\w\$ "
            ;;
        junior|*)
            PS1="(Pre-junior) \u@\h:\w\$ "
            ;;
    esac
fi
