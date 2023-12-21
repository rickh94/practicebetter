{ pkgs, ... }: {
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
  ];

  # https://devenv.sh/scripts/
  scripts.hello.exec = "echo hello from $GREET";

  languages = {
    go.enable = true;
    javascript.enable = true;
  };

  services.caddy = {
    enable = true;
    virtualHosts."pbgo.localhost".extraConfig = ''
      reverse_proxy {
        to :8080
      }
    '';
  };

  certificates = [
    "pbgo.localhost"
  ];

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
