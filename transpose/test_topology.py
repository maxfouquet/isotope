from . import topology


def test_yaml_is_invertable() -> None:
    pass


def test_yaml_load() -> None:
    expected = topology.Topology(services=[
        topology.Service(
            name="a", is_entrypoint=True, response_size=128, script=[])
    ])

    topology_yaml = """services:
- name: a
  isEntrypoint: true
  responseSize: 128
"""
    actual = topology.from_yaml(topology_yaml)
    print(expected.services)
    print(actual.services)

    assert expected == actual
