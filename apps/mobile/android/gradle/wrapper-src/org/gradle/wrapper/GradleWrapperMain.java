package org.gradle.wrapper;

import java.io.*;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.file.*;
import java.util.*;
import java.util.regex.*;
import java.util.zip.*;

/**
 * Small self-contained Gradle wrapper bootstrap used by this generated project.
 * It reads gradle-wrapper.properties, downloads the declared distribution into
 * ~/.gradle/wrapper/dists/ai-fitness-os, unpacks it and delegates to Gradle.
 * The public entry point intentionally matches Gradle's standard wrapper class.
 */
public final class GradleWrapperMain {
    public static void main(String[] args) throws Exception {
        Path jar = Paths.get(GradleWrapperMain.class.getProtectionDomain().getCodeSource().getLocation().toURI());
        Path root = jar.getParent().getParent().getParent();
        Path propsPath = root.resolve("gradle/wrapper/gradle-wrapper.properties");
        Properties props = new Properties();
        try (InputStream in = Files.newInputStream(propsPath)) { props.load(in); }
        String distributionUrl = Objects.requireNonNull(props.getProperty("distributionUrl"), "distributionUrl missing");
        URI uri = URI.create(distributionUrl);
        String fileName = Paths.get(uri.getPath()).getFileName().toString();
        Matcher matcher = Pattern.compile("gradle-(.+?)-(?:bin|all)\\.zip").matcher(fileName);
        if (!matcher.matches()) throw new IllegalArgumentException("Unsupported Gradle distribution: " + fileName);
        String version = matcher.group(1);

        Path cache = Paths.get(System.getProperty("user.home"), ".gradle", "wrapper", "dists", "ai-fitness-os", "gradle-" + version);
        Path gradleHome = cache.resolve("gradle-" + version);
        boolean windows = System.getProperty("os.name").toLowerCase(Locale.ROOT).contains("win");
        Path executable = gradleHome.resolve(windows ? "bin/gradle.bat" : "bin/gradle");

        if (!Files.exists(executable)) {
            Files.createDirectories(cache);
            Path zip = cache.resolve(fileName);
            if (!Files.exists(zip) || Files.size(zip) == 0) download(uri, zip);
            Path staging = cache.resolve("unpack-" + UUID.randomUUID());
            Files.createDirectories(staging);
            try {
                unzip(zip, staging);
                Path unpacked = staging.resolve("gradle-" + version);
                if (!Files.exists(unpacked)) throw new IOException("Gradle archive did not contain gradle-" + version);
                if (Files.exists(gradleHome)) deleteRecursively(gradleHome);
                Files.move(unpacked, gradleHome, StandardCopyOption.ATOMIC_MOVE);
            } finally {
                if (Files.exists(staging)) deleteRecursively(staging);
            }
        }

        if (!windows) executable.toFile().setExecutable(true);
        List<String> command = new ArrayList<>();
        command.add(executable.toAbsolutePath().toString());
        command.addAll(Arrays.asList(args));
        Process process = new ProcessBuilder(command)
            .directory(root.toFile())
            .inheritIO()
            .start();
        System.exit(process.waitFor());
    }

    private static void download(URI uri, Path target) throws Exception {
        System.out.println("Downloading " + uri + " ...");
        HttpClient client = HttpClient.newBuilder().followRedirects(HttpClient.Redirect.ALWAYS).build();
        Path temp = target.resolveSibling(target.getFileName() + ".part");
        HttpRequest request = HttpRequest.newBuilder(uri).GET().build();
        HttpResponse<Path> response = client.send(request, HttpResponse.BodyHandlers.ofFile(temp));
        if (response.statusCode() < 200 || response.statusCode() >= 300) {
            Files.deleteIfExists(temp);
            throw new IOException("Gradle download failed: HTTP " + response.statusCode());
        }
        Files.move(temp, target, StandardCopyOption.REPLACE_EXISTING, StandardCopyOption.ATOMIC_MOVE);
    }

    private static void unzip(Path zip, Path destination) throws IOException {
        try (ZipInputStream input = new ZipInputStream(Files.newInputStream(zip))) {
            ZipEntry entry;
            while ((entry = input.getNextEntry()) != null) {
                Path out = destination.resolve(entry.getName()).normalize();
                if (!out.startsWith(destination)) throw new IOException("Unsafe zip entry: " + entry.getName());
                if (entry.isDirectory()) {
                    Files.createDirectories(out);
                } else {
                    Files.createDirectories(out.getParent());
                    Files.copy(input, out, StandardCopyOption.REPLACE_EXISTING);
                }
                input.closeEntry();
            }
        }
    }

    private static void deleteRecursively(Path path) throws IOException {
        try (var stream = Files.walk(path)) {
            stream.sorted(Comparator.reverseOrder()).forEach(p -> {
                try { Files.deleteIfExists(p); } catch (IOException e) { throw new UncheckedIOException(e); }
            });
        } catch (UncheckedIOException e) {
            throw e.getCause();
        }
    }
}
