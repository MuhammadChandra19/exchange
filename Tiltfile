# Tiltfile

# Common paths that are shared across services
common_only = ["./pkg", "go.work", ".env"]

# Service configuration with opts pattern
# enabled: whether the build is enabled or not
# name: service name (used for docker image and resource naming)
# path: the service path for dockerfile and air config
# only: additional paths specific to this service (combined with common_only)
opts = [
    {
        "enabled": True,
        "name": "eth-btc",
        "path": "matching-service",
        "only": ["./services/matching-service"] + common_only,
        "env_override": {
            "PAIR": "ETH/BTC",
            "PORT": "8081",
            "KAFKA_TOPIC": "eth_btc_orders",
            "KAFKA_GROUP_ID": "eth_btc_group",
        },
    },
    {
        "enabled": True,
        "name": "eth-usdt",
        "path": "matching-service",
        "only": ["./services/matching-service"] + common_only,
        "env_override": {
            "PAIR": "ETH/USDT",
            "PORT": "8082",
            "KAFKA_TOPIC": "eth_usdt_orders",
            "KAFKA_GROUP_ID": "eth_usdt_group",
        },
    },
    {
        "enabled": True,
        "name": "btc-usdt",
        "path": "matching-service",
        "only": ["./services/matching-service"] + common_only,
        "env_override": {
            "PAIR": "BTC/USDT",
            "PORT": "8083",
            "KAFKA_TOPIC": "btc_usdt_orders",
            "KAFKA_GROUP_ID": "btc_usdt_group",
        },
    },
]


# this is required to execute command locally
allow_k8s_contexts(k8s_context())

def run_air():
    """Run services using Air for hot reloading"""
    # Download Air binary
    local_resource(
        "download-air", 
        cmd="go install github.com/cosmtrek/air@v1.40.4", 
        cmd_bat="go install github.com/cosmtrek/air@v1.40.4"
    )
    
    # Load compose files - infrastructure services will start automatically
    docker_compose("./docker-compose.yaml")
    
    # Build and run enabled services with Air
    for opt in opts:
        if not opt["enabled"]:
            continue
            
        override_env = " ".join([
            '{}="{}"'.format(k, v)
            for k, v in opt.get("env_override", {}).items()
        ])

        air_config = os.path.abspath("./dev/docker/{}/.air.toml".format(opt["path"]))

        cmd = 'bash -c "export $(xargs < .env) && {} air -c {}"'.format(
            override_env,
            air_config
        )

        resource_deps = (opt.get("resource_deps") or []) + ["download-air"]
        
        local_resource(
            "[air]" + opt["name"],
            serve_cmd=cmd,
            serve_cmd_bat=cmd,
            allow_parallel=True,
            resource_deps=resource_deps,
            readiness_probe=opt.get("readiness_probe")
        )
    
    # Cleanup old images
    docker_prune_settings()

# def run_docker():
#     """Run services using Docker with live updates"""
#     # Build container images for enabled services
#     for opt in opts:
#         if not opt["enabled"]:
#             continue

#         docker_build(
#             opt["name"],
#             ".",
#             dockerfile="./dev/docker/{}/Dockerfile".format(opt["path"]),
#             only=opt["only"],
#             live_update=[sync(path, "/app/" + path) for path in common_only],
#         )

#     # Load compose files - infrastructure services will start automatically
#     docker_compose("./docker-compose.yaml")
#     docker_compose('./docker-compose-apps.yaml')
    
#     # Cleanup old images
#     docker_prune_settings()

def current_platform():
    """
    Returns current operating system platform
    :return: possible values are "Linux", "Darwin", "Windows", or None
    """
    return str(local("uname -s", command_bat='echo "Windows"', quiet=True)).strip() or None

# Platform detection and execution
platform = current_platform()

# NOTE: if you're switching from run_docker to run_air, make sure you run
# `tilt down` existing resources
if platform == "Linux":
    run_air()
elif platform == "Darwin":
    run_air()
elif platform == "Windows":
    # NOTE: hasn't been tested
    run_air()
else:
    print("unsupported platform: {}.".format(platform))