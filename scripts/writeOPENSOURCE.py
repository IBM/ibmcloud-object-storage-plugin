#!/usr/bin/env python
# ******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Container Service, 5737-D43
# * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************

import yaml

f = open('OPENSOURCE', 'w')

with open("glide.lock", 'r') as stream:
    try:
        data = yaml.load(stream, Loader=yaml.Loader)
        for dep in data["imports"]:
            f.write(dep["name"] + "," + dep["version"] + '\n')
    except yaml.YAMLError as exc:
        print(exc)

f.close()
