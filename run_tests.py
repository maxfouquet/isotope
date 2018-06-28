#!/usr/bin/env python3

import argparse
import logging

from runner import cluster, config as cfg, resources, pipeline

CLUSTER_NAME = 'isotope-cluster'


def main() -> None:
    args = parse_args()

    log_level = getattr(logging, args.log_level)
    logging.basicConfig(level=log_level, format='%(levelname)s\t> %(message)s')

    config = cfg.from_toml_file(args.config_path)

    if config.should_create_cluster:
        cluster.setup(config.cluster_name, config.cluster_zone,
                      config.cluster_version, config.server_machine_type,
                      config.server_disk_size_gb, config.server_num_nodes,
                      config.client_machine_type, config.client_disk_size_gb)

    client_args = ','.join(config.client_args)

    for topology_path in config.topology_paths:
        for environment in config.environments:
            pipeline.run(topology_path, environment, config.server_image,
                         config.client_image, client_args, config.istio_hub,
                         config.istio_tag, config.should_build_istio,
                         config.labels())


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument('config_path', type=str)
    parser.add_argument(
        '--log_level',
        type=str,
        choices=['CRITICAL', 'ERROR', 'WARNING', 'INFO', 'DEBUG'],
        default='INFO')
    return parser.parse_args()


if __name__ == '__main__':
    main()
