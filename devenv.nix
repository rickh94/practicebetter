{ pkgs
, config
, ...
}: {
  packages = with pkgs; [
    git
    bun
    air
    litestream
    atlas
    sqlfluff
    sqlc
    nodePackages.eslint
    nodePackages.prettier
    mkcert
    caddy
    golangci-lint
    pre-commit
    typescript
  ];

  # https://devenv.sh/scripts/
  scripts.hello.exec = "echo hello from $GREET";

  languages = {
    go.enable = true;
    javascript.enable = true;
  };

  certificates = [
    "pbgo.localhost"
  ];

  services.caddy = {
    enable = true;
    virtualHosts."pbgo.localhost".extraConfig = ''
      request_body * {
        max_size 1000MB
      }
      reverse_proxy {
        to :8080
      }
      tls ${config.env.DEVENV_STATE}/mkcert/pbgo.localhost.pem ${config.env.DEVENV_STATE}/mkcert/pbgo.localhost-key.pem
    '';
  };

  services.redis = {
    enable = true;
    port = 6372;
  };

  # See full reference at https://devenv.sh/reference/options/
  processes = {
    air.exec = "air";
    litestream.exec = "${pkgs.litestream}/bin/litestream replicate -config ${./litestream.dev.yml}";
  };

  scripts = {
    # install bun, templ, and staticfiles
    install.exec = "bun install && go install github.com/a-h/templ/cmd/templ@latest";
  };
}
