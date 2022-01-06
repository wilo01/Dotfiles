# Dotfiles and fresh install guide

My dev config.

## **Step 1 – Install _[Homebrew](http://brew.sh/)_**

[Open a terminal window](http://www.youtube.com/watch?v=zw7Nd67_aFw).

> Homebrew [...] simplifies the installation of software on the Mac OS X operating system.

– [Homebrew – Wikipedia](http://en.wikipedia.org/wiki/Homebrew_%28package_management_software%29)

**Copy & paste the following** into the terminal window and **hit `Return`**.

```jsx
/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
brew doctor
```

You will be requested to install the _Command Line Developer Tools_ from _Apple_. **Confirm by clicking _Install_**. After the installation finished, continue installing _Homebrew_ by **hitting `Return`** again.

## **Step 2 – Install _zsh_**

**Copy & paste the following** into the terminal window and **hit `Return`**.

```jsx
brew install zsh
```

**You can use _zsh_ now.**

## **_Optional_ step 3 – default shell**

To set _zsh_ as your default shell, execute the following.

```jsx
sudo sh -c "echo $(which zsh) >> /etc/shells" && chsh -s $(which zsh)
```

3. Install OhMyZsh

```jsx
sh -c "$(curl -fsSL [https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh](https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh))"
```

```jsx
ZSH_THEME="powerlevel10k/powerlevel10k"
POWERLEVEL9K_MODE="nerdfont-complete"
plugins=(git zsh-autosuggestions zsh-syntax-highlighting)
```

4. Install PowerLeve10K theme && Plugins for autosuggestion and syntax highlighting

```jsx
git clone https://github.com/romkatv/powerlevel10k.git $ZSH_CUSTOM/themes/powerlevel10k
&&
git clone https://github.com/zsh-users/zsh-autosuggestions ~/.oh-my-zsh/custom/plugins/zsh-autosuggestions
&&
git clone https://github.com/zsh-users/zsh-syntax-highlighting.git ~/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting
```

## **Step 4 – Install fonts**

[ryanoasis/nerd-fonts](https://github.com/ryanoasis/nerd-fonts/tree/master/patched-fonts/FiraMono)

[Install Nerd Fonts via Homebrew](https://gist.github.com/davidteren/898f2dcccd42d9f8680ec69a3a5d350e)

[Nerd fonts raw](https://gist.githubusercontent.com/davidteren/898f2dcccd42d9f8680ec69a3a5d350e/raw/f0ad2522b87caf640c9d2d5a70ddd979d8d92458/nerd_fonts.sh)

- **Homebrew**

```jsx
brew tap homebrew/cask-fonts &&
brew install --cask font-<FONT NAME>-nerd-font

brew tap homebrew/cask-fonts &&
brew install --cask font-meslo-lg-nerd-font

# Nerd Fonts for your IDE
# https://www.nerdfonts.com/font-downloads

**brew tap homebrew/cask-fonts && brew install --cask font-FiraMono-nerd-font**
**brew tap homebrew/cask-fonts && brew install --cask font-meslo-nerd-font**
```

- **:// curl**

```jsx
curl -fLo "<FONT NAME> Nerd Font Complete.otf" \
    https://github.com/ryanoasis/nerd-fonts/raw/master/patched-fonts/<FONT_PATH>/complete/<FONT_NAME>%20Nerd%20Font%20Complete.otf
```

- **Bash Install Script**

```jsx
./install.sh <FontName>
```

6. Edit ~.zshrc and add:

## **Step 6 – Download config**

Install:
Git Clone in user directory /home/$(whoami)

```jsx
git clone git@github.com:wilo01/Dotfiles.git
```

To install nvim plugins open nvim and type:
:PlugInstall
:PlugUpdate

Commands Used 💻 Installing ZSH shell for linux distros:

## **Step 6 – Download config**

```jsx
brew install --cask visual-studio-code

defaults write NSGlobalDomain KeyRepeat -int 0
defaults write NSGlobalDomain ApplePressAndHoldEnabled -bool false
defaults write -g ApplePressAndHoldEnabled -bool false

reboot
```

## **Step 7 – Git config with ssh**

[Downloads](https://git-scm.com/downloads)

**SSH Github key !**

```jsx
git config --global user.name "wilo01"
git config --global user.email "wilk9496@gmail.com"
git config --global core.editor "nvim"

ssh-keygen -t ed25519 -C "wilk9496@gmail.com"
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519
cat ~/.ssh/id_ed25519.pub

copy ssh key to github
reboot if problems
```

**SSH without password**

```jsx
// ssh copy id
ssh-copy-id pi@ip-adress or ssh-copy-id pi@domain-name.local

ssh-copy-id pi@10.1.1.206
ssh-copy-id pi@10.1.1.208

// In sshd_config change line PsswordAuthentication no
sudo nvim /etc/ssh/sshd_config

PsswordAuthentication no

sudo service ssh restart
```

[Connecting to GitHub with SSH](https://docs.github.com/en/github/authenticating-to-github/connecting-to-github-with-ssh)
