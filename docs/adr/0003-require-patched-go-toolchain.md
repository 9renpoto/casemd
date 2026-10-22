# Require the Patched Go Toolchain

The eMudhra certificate remediation requires checkout builds to use a Go release containing the x509 distrust fix, so `go.mod` sets Go 1.27.3 as the minimum toolchain instead of leaving the vulnerable Go 1.27.1 available to contributors.
Go 1.27.3 and later are supported; the Go command automatically selects a newer toolchain when necessary.
This raises the minimum from Go 1.26 and contributors must use or allow automatic selection of Go 1.27.3 or later.

<details><summary>日本語</summary>

eMudhra 証明書への対応として、checkout からのビルドでも x509 の distrust 修正を含む Go release を使う必要があるため、`go.mod` の最低 toolchain を Go 1.27.3 にします。
これにより、脆弱な Go 1.27.1 が開発者のビルドで使われることを防ぎます。
Go 1.27.3 以降をサポートし、必要な場合は Go コマンドが新しい toolchain を自動選択します。
最低バージョンを Go 1.26 から引き上げるため、開発者は Go 1.27.3 以降を利用するか、自動選択を許可する必要があります。

</details>
