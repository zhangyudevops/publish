node {
    checkout scm

    stage('build package') {
        echo "Build Package Stage"
        sh "/usr/local/go/bin/go env -w GO111MODULE=on"
        sh "/usr/local/go/bin/go env -w GOPROXY=https://goproxy.cn,direct"
        sh 'CGO_ENABLED=0 /usr/local/go/bin/go build -o ./main -v .'
    }

    stage('deploy application') {
        echo "Deploy Application Stage"
        ansiColor('xterm') {
            ansiblePlaybook(
                playbook: './playbook.yaml',
                colorized: true
            )}
    }

    stage("清理Jenkins工作空间") {
        deleteDir()
    }
}