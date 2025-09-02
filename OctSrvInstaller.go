package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// 改這裡
	MODPACK_URL  = "https://mediafilez.forgecdn.net/files/6936/279/ATM10%20To%20the%20Sky-1.2.1.zip"
	MODPACK_NAME = "[01.014.00] atm 10 sky"

	// 以下通常不用改
	PRISM_URL = "https://github.com/PrismLauncher/PrismLauncher/releases/download/9.4/PrismLauncher-Windows-MinGW-w64-Portable-9.4.zip"
	ICON_URL  = "https://raw.githubusercontent.com/octoberserver/octsrvmodpack/refs/heads/main/october.ico"
)

func main() {
	appdata := os.Getenv("APPDATA")
	prismDir := filepath.Join(appdata, "oct-launcher")
	instancesDir := filepath.Join(prismDir, "instances")
	iconPath := filepath.Join(prismDir, "october.ico")

	// 檢查 Prism
	if _, err := os.Stat(prismDir); os.IsNotExist(err) {
		fmt.Println("未偵測到十月模組包啟動器，開始下載...")
		os.MkdirAll(prismDir, 0755)

		prismZip := filepath.Join(filepath.Dir(prismDir), "PrismLauncher.zip")
		download(PRISM_URL, prismZip)
		unzip(prismZip, prismDir)
		os.Remove(prismZip)

		// 下載圖示
		download(ICON_URL, iconPath)

		// 建立捷徑
		prismExe := filepath.Join(prismDir, "prismlauncher.exe")
		desktop := filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
		err = createShortcut(desktop, "Oct Mod Launcher", prismExe, iconPath)
		if err != nil {
			fmt.Println("建立捷徑失敗:", err)
		}
	} else {
		fmt.Println("已偵測到十月模組包啟動器，跳過安裝與捷徑建立。")
	}

	// 安裝模組包
	modpackDir := filepath.Join(instancesDir, MODPACK_NAME)
	if _, err := os.Stat(modpackDir); os.IsNotExist(err) {
		fmt.Println("未偵測到模組包，開始下載...")
		os.MkdirAll(instancesDir, 0755)

		modpackZip := filepath.Join(filepath.Dir(prismDir), MODPACK_NAME+".zip")
		download(MODPACK_URL, modpackZip)

		prismExe := filepath.Join(prismDir, "prismlauncher.exe")
		cmd := exec.Command(prismExe,
			"-d", prismDir,
			"-I", modpackZip,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()

		if err == nil {
			fmt.Println("模組包安裝完成！")
		} else {
			fmt.Println("模組包安裝失敗:", err)
		}
		os.Remove(modpackZip)
	} else {
		fmt.Println("模組包已存在，跳過下載。")
	}

	fmt.Println("安裝流程完成！")
}

func download(url string, dest string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("下載完成:", dest)
}

func unzip(src string, dest string) {
	r, err := zip.OpenReader(src)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			panic(err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		rc, err := f.Open()
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(outFile, rc)
		if err != nil {
			panic(err)
		}

		outFile.Close()
		rc.Close()
	}
	fmt.Println("解壓完成:", dest)
}

func createShortcut(desktop, name, targetPath, iconPath string) error {
	shortcutPath := filepath.Join(desktop, name+".lnk")

	// PowerShell script 內容
	script := fmt.Sprintf(`$s = (New-Object -ComObject WScript.Shell).CreateShortcut('%s');
$s.TargetPath = '%s';
$s.IconLocation = '%s,0';
$s.Save();`, shortcutPath, targetPath, iconPath)

	// 存成暫存 ps1
	tmpfile, err := os.CreateTemp("", "createshortcut*.ps1")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(script)); err != nil {
		return err
	}
	tmpfile.Close()

	// 呼叫 PowerShell 執行
	cmd := exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", tmpfile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	if err != nil {
		return fmt.Errorf("PowerShell 建立捷徑失敗: %w", err)
	}
	fmt.Println("捷徑已建立:", shortcutPath)
	return nil
}
