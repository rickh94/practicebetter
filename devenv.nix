{ pkgs, ... }: {
  packages = with pkgs; [
    git
    bun
    air
    litestream
    atlas
    sqlfluff
    sqlc
  ];

  # https://devenv.sh/scripts/
  scripts.hello.exec = "echo hello from $GREET";

  languages = {
    go.enable = true;
    javascript.enable = true;
  };

  certificates = [
    "musiclib.localhost"
  ];

  services.caddy = {
    enable = true;
    virtualHosts."musiclib.localhost".extraConfig = ''
      reverse_proxy {
        to :8080
      }
    '';
  };

  services.redis = {
    enable = true;
    port = 6375;
  };

  # See full reference at https://devenv.sh/reference/options/
  processes = {
    air.exec = "air";
    litestream.exec = "${pkgs.litestream}/bin/litestream replicate -config ${./litestream.dev.yml}";
  };

  scripts = {
    tw.exec = "bun run watch";

    # install bun, templ, and staticfiles
    install.exec = "bun install && go install github.com/a-h/templ/cmd/templ@latest";
  };
}
