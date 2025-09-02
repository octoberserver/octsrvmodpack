import java.io.*;
import java.net.URL;
import java.nio.file.*;

public class OctSrvInstaller {
    //改這裡改模組包網址和名稱
    private static final String MODPACK_URL = "https://mediafilez.forgecdn.net/files/6936/279/ATM10%20To%20the%20Sky-1.2.1.zip";
    private static final String MODPACK_NAME = "[01.014.00] atm 10 sky";

    // 以下通常不需要改
    private static final String APPDATA = System.getenv("APPDATA");
    private static final Path PRISM_DIR = Paths.get(APPDATA, "oct-launcher");
    private static final Path INSTANCES_DIR = PRISM_DIR.resolve("instances");
    private static final Path ICON_PATH = PRISM_DIR.resolve("october.ico");
    private static final String PRISM_URL = "https://github.com/PrismLauncher/PrismLauncher/releases/download/9.4/PrismLauncher-Windows-MinGW-w64-Portable-9.4.zip";
    private static final String ICON_URL = "https://raw.githubusercontent.com/octoberserver/octsrvmodpack/refs/heads/main/october.ico";

    public static void main(String[] args) {
        try {
            boolean prismInstalled = Files.exists(PRISM_DIR);

            // 1️ 安裝 Prism
            if (!prismInstalled) {
                System.out.println("未偵測到十月模組包啟動器，開始下載...");
                Files.createDirectories(PRISM_DIR);

                Path prismZip = PRISM_DIR.resolveSibling("PrismLauncher.zip");
                download(PRISM_URL, prismZip);
                unzip(prismZip, PRISM_DIR);
                Files.deleteIfExists(prismZip);

                // 下載圖示
                download(ICON_URL, ICON_PATH);

                // 建立捷徑
                Path prismExe = PRISM_DIR.resolve("prismlauncher.exe");
                createShortcut("十月模組包啟動器 Oct Mod Launcher", prismExe.toString(), ICON_PATH.toString());
            } else {
                System.out.println("已偵測到十月模組包啟動器，跳過安裝與捷徑建立。");
            }

            // 2 安裝模組包
            Path modpackDir = INSTANCES_DIR.resolve(MODPACK_NAME);
            if (!Files.exists(modpackDir)) {
                System.out.println("未偵測到模組包，開始下載...");
                Files.createDirectories(INSTANCES_DIR);

                Path modpackZip = PRISM_DIR.resolveSibling(MODPACK_NAME+".zip");
                download(MODPACK_URL, modpackZip);

                // 呼叫 Prism --import
                Path prismExe = PRISM_DIR.resolve("prismlauncher.exe");
                ProcessBuilder pb = new ProcessBuilder(
                        prismExe.toString(),
                        "-d", PRISM_DIR.toString(),
                        "-I", modpackZip.toString()
                );
                pb.inheritIO();
                int exitCode = pb.start().waitFor();

                if (exitCode == 0) {
                    System.out.println("模組包安裝完成！");
                } else {
                    System.err.println("模組包安裝失敗，Prism 回傳代碼: " + exitCode);
                }

                Files.deleteIfExists(modpackZip);
            } else {
                System.out.println("模組包已存在，跳過下載。");
            }

            System.out.println("安裝流程完成！");
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    private static void download(String urlStr, Path dest) throws IOException {
        try (InputStream in = new URL(urlStr).openStream()) {
            Files.copy(in, dest, StandardCopyOption.REPLACE_EXISTING);
        }
        System.out.println("下載完成: " + dest);
    }

    private static void unzip(Path zipFile, Path destDir) throws IOException {
        try (java.util.zip.ZipInputStream zis = new java.util.zip.ZipInputStream(Files.newInputStream(zipFile))) {
            java.util.zip.ZipEntry entry;
            while ((entry = zis.getNextEntry()) != null) {
                Path newFile = destDir.resolve(entry.getName());
                if (entry.isDirectory()) {
                    Files.createDirectories(newFile);
                } else {
                    Files.createDirectories(newFile.getParent());
                    Files.copy(zis, newFile, StandardCopyOption.REPLACE_EXISTING);
                }
                zis.closeEntry();
            }
        }
        System.out.println("解壓完成: " + destDir);
    }

private static void createShortcut(String name, String targetPath, String iconPath) throws IOException {
    String desktop = System.getProperty("user.home") + "\\Desktop";
    String startMenu = System.getenv("APPDATA") + "\\Microsoft\\Windows\\Start Menu\\Programs";

    String psScript = String.format(
        "$WshShell = New-Object -ComObject WScript.Shell; " +
        "$Shortcut = $WshShell.CreateShortcut('%s\\%s.lnk'); " +
        "$Shortcut.TargetPath = '%s'; " +
        "$Shortcut.IconLocation = '%s'; " +
        "$Shortcut.Save();",
        desktop, name, targetPath, iconPath
    );
    new ProcessBuilder("powershell", "-Command", psScript).start();

    psScript = String.format(
        "$WshShell = New-Object -ComObject WScript.Shell; " +
        "$Shortcut = $WshShell.CreateShortcut('%s\\%s.lnk'); " +
        "$Shortcut.TargetPath = '%s'; " +
        "$Shortcut.IconLocation = '%s'; " +
        "$Shortcut.Save();",
        startMenu, name, targetPath, iconPath
    );
    new ProcessBuilder("powershell", "-Command", psScript).start();

    System.out.println("捷徑已建立: " + name);
}
}
