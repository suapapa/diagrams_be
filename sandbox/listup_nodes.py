import glob
import importlib
import inspect
import json
import pkgutil
import os
import sys
import sysconfig

pkg_dir = os.path.join(sysconfig.get_paths()["purelib"], "diagrams")
rsc_dir = os.path.join(sysconfig.get_paths()["purelib"], "resources")

providers_dirs = glob.glob(os.path.join(pkg_dir, "*"))
providers_dirs = list(filter(lambda d : (os.path.isdir(d) and not os.path.basename(d).startswith("__")), providers_dirs))

providers = list(map(lambda d: os.path.basename(d), providers_dirs))

for pkg_dir in providers_dirs:
    for (_, node_name, _) in pkgutil.iter_modules([pkg_dir]):
        provider = os.path.basename(pkg_dir)
        try:
            importlib.import_module("diagrams." + provider + "." + node_name, __package__)
        except: # TODO: handle exception
            pass

# [{"pkg":"diagrams.oci.storage", "node":"StorageGatewayWhite", "icon":"storage-gateway-white.png"}, ...]
json_list = []
for m in sys.modules.keys():
    if not m.startswith("diagrams"):
        pass
    for n, c in inspect.getmembers(sys.modules[m], inspect.isclass):
        try:
            if c._icon:
                m = c.__module__
                icon = c._icon
                json_list.append({"module":m, "node":n, "icon":icon})
        except AttributeError: # if it has no icon, don't list up.
            pass

# print json_list as json
print(json.dumps(json_list, indent=2))

# all done