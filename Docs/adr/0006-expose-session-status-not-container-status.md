# APIにはSessionの状態だけを公開する

Session取得APIは `running` または `stopped` のSession状態を返し、実行基盤の詳細な状態名を公開契約に含めない。現在の実行手段で実行中なら `running`、それ以外なら `stopped` へ対応付け、詳細な内部状態は診断情報として扱う。これにより実行手段を変更してもAPIの状態表現を維持する。
