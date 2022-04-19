package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"strings"
	"time"
)

func init() {
	//设置中文字体
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		//楷体:simkai.ttf
		//黑体:simhei.ttf
		if strings.Contains(path, "simkai.ttf") {
			err := os.Setenv("FYNE_FONT", path)
			if err != nil {
				log.Println("设置字体全局变量异常")
			}
			break
		}
	}
}

var A fyne.App
var W fyne.Window
var Client *ssh.Client

//var Session *ssh.Session

const SshPassword = "Password"

//const SshKey = ""

func main() {
	A = app.New()
	A.SetIcon(theme.FyneLogo())
	logLifecycle()
	W = A.NewWindow("Crust工具箱")
	W.SetMainMenu(makeMenu())
	W.SetMaster()

	tabs := container.NewAppTabs(
		container.NewTabItem("远程命令", tab1()),
	)

	tabs.SetTabLocation(container.TabLocationLeading)
	W.SetContent(tabs)
	W.Resize(fyne.NewSize(800, 600))
	W.ShowAndRun()
	defer func() {
		if err := os.Unsetenv("FYNE_FONT"); err != nil {
			log.Println("取消字体全局变量异常")
		}
	}()
}

func logLifecycle() {
	// 生命周期记录
	A.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	A.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	A.Lifecycle().SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
	})
	A.Lifecycle().SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}

func makeMenu() *fyne.MainMenu {
	themeLight := fyne.NewMenuItem("Light", func() {
		setTheme("Light")
	})
	themeDark := fyne.NewMenuItem("Dark", func() {
		setTheme("Dark")
	})
	setting := fyne.NewMenuItem("设置", nil)
	menuMain := fyne.NewMenu("菜单", setting)
	menuSet := fyne.NewMenu("主题", themeLight, themeDark)
	mainMenu := fyne.NewMainMenu(menuMain, menuSet)
	return mainMenu
}

func tab1() *fyne.Container {
	// 创建第一个tab
	mainV := container.NewVBox()
	//ipL := widget.NewLabel("主机")
	ipE := widget.NewEntry()
	//portL := widget.NewLabel("端口")
	portE := widget.NewEntry()
	//userL := widget.NewLabel("用户名")
	userE := widget.NewEntry()
	//passL := widget.NewLabel("密码")
	command := widget.NewEntry()
	passE := widget.NewPasswordEntry()
	ipE.SetPlaceHolder("IP地址/192.168.1.1")
	portE.SetPlaceHolder("端口号/22")
	userE.SetPlaceHolder("用户名/user")
	passE.SetPlaceHolder("密码/password")
	// 调试
	ipE.SetText("192.168.181.131")
	portE.SetText("22")
	userE.SetText("root")
	passE.SetText("123456")

	terminal := widget.NewMultiLineEntry()
	terminal.Wrapping = fyne.TextWrapWord
	//terminal.Disable()

	mainV.Add(ipE)
	mainV.Add(portE)
	mainV.Add(userE)
	mainV.Add(passE)
	open := widget.NewButton("连接", func() {
		go manageSsh(ipE, portE, userE, passE, terminal, command)
	})
	exit := widget.NewButton("关闭", func() {
		err := Client.Close()
		if err != nil {
			log.Println("关闭Client异常", err)
		}
		terminalUp(terminal, "已断开连接！")
	})
	spt := container.NewHSplit(open, exit)
	mainV.Add(spt)
	bsp := container.NewVSplit(terminal, command)
	bsp.Offset = 1
	vsp := container.NewVSplit(mainV, bsp)
	vsp.Offset = 0
	return container.NewMax(vsp)
}

func setTheme(ty string) {
	// 设置主题模式
	if ty == "Light" {
		A.Settings().SetTheme(theme.LightTheme())
	} else {
		A.Settings().SetTheme(theme.DarkTheme())
	}
}

func connect(sshType, host, port, user, password, keyPath string) (err error) {
	//配置连接
	config := &ssh.ClientConfig{
		Timeout:         time.Second * 3,
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //不够安全
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
	}
	if sshType == "Password" {
		config.Auth = []ssh.AuthMethod{ssh.Password(password)}
	} else {
		//config.Auth = []ssh.AuthMethod{publicKeyAuthFunc(sshKeyPath)}
		return nil
	}

	//dial 获取ssh client
	addr := fmt.Sprintf("%s:%s", host, port)
	Client, err = ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("创建ssh client 失败", err)
		return err
	}
	//defer Client.Close()

	return nil

}

func execute(command string) string {
	//执行远程命令
	Session, err := Client.NewSession()
	if err != nil {
		log.Fatal("创建ssh session 失败", err)
		return "连接失败"
	}
	defer Session.Close()

	combo, err := Session.Output(command)
	if err != nil {
		log.Fatal("远程执行cmd 失败", err)
		return "不支持此命令"
	}
	str := string(combo)
	// 去除命令执行后结尾的换行符
	str = strings.TrimSuffix(str, "\n")
	fmt.Println(str)
	return str
}

func manageSsh(host, port, user, password, terminal, command *widget.Entry) {
	//连接的生命周期管理

	terminal.SetText(fmt.Sprintf("正在连接 %s:%s", host.Text, port.Text))

	err := connect(SshPassword, host.Text, port.Text, user.Text, password.Text, "")
	if err != nil {
		dialog.ShowError(errors.New(fmt.Sprintf("%s", err)), W)
		return
	} else {
		terminalUp(terminal, "连接成功！")
	}
	command.OnSubmitted = func(s string) {
		terminalUp(terminal, s)
		terminalUp(terminal, execute(s))
		command.SetText("")
	}

	time.Sleep(time.Second * 60 * 5)
	terminalUp(terminal, "连接长时间无人使用，关闭连接")

	defer Client.Close()

}

func terminalUp(t *widget.Entry, mes string) {
	// 更新终端显示界面
	t.SetText(fmt.Sprintf("%s\n%s", t.Text, mes))
}
